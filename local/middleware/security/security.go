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

// RateLimiter stores rate limiting information
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	// Use graceful shutdown for cleanup goroutine
	shutdownManager := graceful.GetManager()
	shutdownManager.RunGoroutine(func(ctx context.Context) {
		rl.cleanupWithContext(ctx)
	})

	return rl
}

// cleanup removes old entries from the rate limiter
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

// SecurityMiddleware provides comprehensive security middleware
type SecurityMiddleware struct {
	rateLimiter *RateLimiter
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware() *SecurityMiddleware {
	return &SecurityMiddleware{
		rateLimiter: NewRateLimiter(),
	}
}

// SecurityHeaders adds security headers to responses
func (sm *SecurityMiddleware) SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Set("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Set("X-XSS-Protection", "1; mode=block")

		// Prevent referrer leakage
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")

		// Permissions Policy
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")

		return c.Next()
	}
}

// RateLimit implements rate limiting for API endpoints
func (sm *SecurityMiddleware) RateLimit(maxRequests int, duration time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		sm.rateLimiter.mutex.Lock()
		defer sm.rateLimiter.mutex.Unlock()

		now := time.Now()
		requests := sm.rateLimiter.requests[key]

		// Remove requests older than duration
		filtered := make([]time.Time, 0, len(requests))
		for _, t := range requests {
			if now.Sub(t) < duration {
				filtered = append(filtered, t)
			}
		}

		// Check if limit is exceeded
		if len(filtered) >= maxRequests {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Rate limit exceeded",
				"retry_after": duration.Seconds(),
			})
		}

		// Add current request
		filtered = append(filtered, now)
		sm.rateLimiter.requests[key] = filtered

		return c.Next()
	}
}

// AuthRateLimit implements stricter rate limiting for authentication endpoints
func (sm *SecurityMiddleware) AuthRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		userAgent := c.Get("User-Agent")
		key := fmt.Sprintf("%s:%s", ip, userAgent)

		sm.rateLimiter.mutex.Lock()
		defer sm.rateLimiter.mutex.Unlock()

		now := time.Now()
		requests := sm.rateLimiter.requests[key]

		// Remove requests older than 15 minutes
		filtered := make([]time.Time, 0, len(requests))
		for _, t := range requests {
			if now.Sub(t) < 15*time.Minute {
				filtered = append(filtered, t)
			}
		}

		// Check if limit is exceeded (5 requests per 15 minutes for auth)
		if len(filtered) >= 5 {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too many authentication attempts",
				"retry_after": 900, // 15 minutes
			})
		}

		// Add current request
		filtered = append(filtered, now)
		sm.rateLimiter.requests[key] = filtered

		return c.Next()
	}
}

// InputSanitization sanitizes user input to prevent XSS and injection attacks
func (sm *SecurityMiddleware) InputSanitization() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Sanitize query parameters
		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			sanitized := sanitizeInput(string(value))
			c.Request().URI().QueryArgs().Set(string(key), sanitized)
		})

		// Store original body for processing
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			body := c.Body()
			if len(body) > 0 {
				// Basic sanitization - remove potentially dangerous patterns
				sanitized := sanitizeInput(string(body))
				c.Request().SetBodyString(sanitized)
			}
		}

		return c.Next()
	}
}

// sanitizeInput removes potentially dangerous patterns from input
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

// ValidateContentType ensures only expected content types are accepted
func (sm *SecurityMiddleware) ValidateContentType(allowedTypes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			contentType := c.Get("Content-Type")
			if contentType == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Content-Type header is required",
				})
			}

			// Check if content type is allowed
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

// ValidateUserAgent blocks requests with suspicious or missing user agents
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
		"curl/7.0", // Very old curl versions
		"wget/1.0", // Very old wget versions
	}

	return func(c *fiber.Ctx) error {
		userAgent := strings.ToLower(c.Get("User-Agent"))

		// Block empty user agents
		if userAgent == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User-Agent header is required",
			})
		}

		// Block suspicious user agents
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

// RequestSizeLimit limits the size of incoming requests
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

// LogSecurityEvents logs security-related events
func (sm *SecurityMiddleware) LogSecurityEvents() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Log suspicious activity
		status := c.Response().StatusCode()
		if status == 401 || status == 403 || status == 429 {
			duration := time.Since(start)
			// In a real implementation, you would send this to your logging system
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

// TimeoutMiddleware adds request timeout
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
