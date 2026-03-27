package security

import (
	"acc-server-manager/local/utl/graceful"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	shutdownManager := graceful.GetManager()
	shutdownManager.RunGoroutine(func(ctx context.Context) {
		rl.cleanupWithContext(ctx)
	})

	return rl
}

func (rl *RateLimiter) cleanupWithContext(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mutex.Lock()
			now := time.Now()
			for key, times := range rl.requests {
				filtered := make([]time.Time, 0, len(times))
				for _, t := range times {
					if now.Sub(t) < time.Hour {
						filtered = append(filtered, t)
					}
				}
				if len(filtered) == 0 {
					delete(rl.requests, key)
				} else {
					rl.requests[key] = filtered
				}
			}
			rl.mutex.Unlock()
		}
	}
}

type SecurityMiddleware struct {
	rateLimiter *RateLimiter
}

func NewSecurityMiddleware() *SecurityMiddleware {
	return &SecurityMiddleware{
		rateLimiter: NewRateLimiter(),
	}
}

func (sm *SecurityMiddleware) SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")

		return c.Next()
	}
}

func (sm *SecurityMiddleware) RateLimit(maxRequests int, duration time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		sm.rateLimiter.mutex.Lock()
		defer sm.rateLimiter.mutex.Unlock()

		now := time.Now()
		requests := sm.rateLimiter.requests[key]

		filtered := make([]time.Time, 0, len(requests))
		for _, t := range requests {
			if now.Sub(t) < duration {
				filtered = append(filtered, t)
			}
		}

		if len(filtered) >= maxRequests {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Rate limit exceeded",
				"retry_after": duration.Seconds(),
			})
		}

		filtered = append(filtered, now)
		sm.rateLimiter.requests[key] = filtered

		return c.Next()
	}
}

func (sm *SecurityMiddleware) AuthRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		userAgent := c.Get("User-Agent")
		key := fmt.Sprintf("%s:%s", ip, userAgent)

		sm.rateLimiter.mutex.Lock()
		defer sm.rateLimiter.mutex.Unlock()

		now := time.Now()
		requests := sm.rateLimiter.requests[key]

		filtered := make([]time.Time, 0, len(requests))
		for _, t := range requests {
			if now.Sub(t) < 15*time.Minute {
				filtered = append(filtered, t)
			}
		}

		if len(filtered) >= 5 {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too many authentication attempts",
				"retry_after": 900, // 15 minutes
			})
		}

		filtered = append(filtered, now)
		sm.rateLimiter.requests[key] = filtered

		return c.Next()
	}
}

func (sm *SecurityMiddleware) InputSanitization() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			sanitized := sanitizeInput(string(value))
			c.Request().URI().QueryArgs().Set(string(key), sanitized)
		})

		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			body := c.Body()
			if len(body) > 0 {
				sanitized := sanitizeInput(string(body))
				c.Request().SetBodyString(sanitized)
			}
		}

		return c.Next()
	}
}

func sanitizeInput(input string) string {
	dangerous := []string{
		"<script",
		"</script>",
		"javascript:",
		"vbscript:",
		"data:text/html",
		"data:application",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
		"onchange=",
		"onsubmit=",
		"onkeydown=",
		"onkeyup=",
		"<iframe",
		"<object",
		"<embed",
		"<link",
		"<meta",
		"<style",
		"<form",
		"<input",
		"<button",
		"<svg",
		"<math",
		"expression(",
		"@import",
		"url(",
		"\\x",
		"\\u",
		"&#x",
		"&#",
	}

	result := input
	lowerInput := strings.ToLower(input)

	for _, pattern := range dangerous {
		if strings.Contains(lowerInput, pattern) {
			return ""
		}
	}

	if strings.Contains(result, "\x00") {
		return ""
	}

	if len(strings.TrimSpace(result)) == 0 && len(input) > 0 {
		return ""
	}

	return result
}

func (sm *SecurityMiddleware) ValidateContentType(allowedTypes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			contentType := c.Get("Content-Type")
			if contentType == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Content-Type header is required",
				})
			}

			allowed := false
			for _, allowedType := range allowedTypes {
				if strings.Contains(contentType, allowedType) {
					allowed = true
					break
				}
			}

			if !allowed {
				return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
					"error": "Unsupported content type",
				})
			}
		}

		return c.Next()
	}
}

func (sm *SecurityMiddleware) ValidateUserAgent() fiber.Handler {
	suspiciousAgents := []string{
		"sqlmap",
		"nikto",
		"nmap",
		"masscan",
		"gobuster",
		"dirb",
		"dirbuster",
		"wpscan",
		"curl/7.0",
		"wget/1.0",
	}

	return func(c *fiber.Ctx) error {
		userAgent := strings.ToLower(c.Get("User-Agent"))

		if userAgent == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User-Agent header is required",
			})
		}

		for _, suspicious := range suspiciousAgents {
			if strings.Contains(userAgent, suspicious) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Access denied",
				})
			}
		}

		return c.Next()
	}
}

func (sm *SecurityMiddleware) RequestSizeLimit(maxSize int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			contentLength := c.Request().Header.ContentLength()
			if contentLength > maxSize {
				return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
					"error":    "Request too large",
					"max_size": maxSize,
				})
			}
		}

		return c.Next()
	}
}

func (sm *SecurityMiddleware) LogSecurityEvents() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		status := c.Response().StatusCode()
		if status == 401 || status == 403 || status == 429 {
			duration := time.Since(start)
			fmt.Printf("[SECURITY] %s %s %s %d %v %s\n",
				time.Now().Format(time.RFC3339),
				c.IP(),
				c.Method(),
				status,
				duration,
				c.Path(),
			)
		}

		return err
	}
}

func (sm *SecurityMiddleware) TimeoutMiddleware(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), timeout)
		defer cancel()

		c.SetUserContext(ctx)

		return c.Next()
	}
}

func (sm *SecurityMiddleware) RequestContextTimeout(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), timeout)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- c.Next()
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{
				"error": "Request timeout",
			})
		}
	}
}
