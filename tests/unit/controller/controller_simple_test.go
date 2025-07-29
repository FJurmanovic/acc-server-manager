package controller

import (
	"acc-server-manager/local/model"
	"acc-server-manager/tests"
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TestController_JSONParsing_Success(t *testing.T) {
	// Setup environment and test helper
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Test basic JSON parsing functionality
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		var data map[string]interface{}
		if err := c.BodyParser(&data); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		return c.JSON(data)
	})

	// Prepare test data
	testData := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}
	bodyBytes, err := json.Marshal(testData)
	tests.AssertNoError(t, err)

	// Create request
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)
	err = json.Unmarshal(body, &response)
	tests.AssertNoError(t, err)

	// Verify response
	tests.AssertEqual(t, "test", response["name"])
	tests.AssertEqual(t, float64(123), response["value"]) // JSON numbers are float64
}

func TestController_JSONParsing_InvalidJSON(t *testing.T) {
	// Setup environment and test helper
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Test handling of invalid JSON
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		var data map[string]interface{}
		if err := c.BodyParser(&data); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		return c.JSON(data)
	})

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 400, resp.StatusCode)

	// Parse error response
	var response map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)
	err = json.Unmarshal(body, &response)
	tests.AssertNoError(t, err)

	// Verify error response
	tests.AssertEqual(t, "Invalid JSON", response["error"])
}

func TestController_UUIDValidation_Success(t *testing.T) {
	// Setup environment and test helper
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Test UUID parameter validation
	app := fiber.New()

	app.Get("/test/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Validate UUID
		if _, err := uuid.Parse(id); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid UUID"})
		}

		return c.JSON(fiber.Map{"id": id, "valid": true})
	})

	// Create request with valid UUID
	validUUID := uuid.New().String()
	req := httptest.NewRequest("GET", "/test/"+validUUID, nil)

	// Execute request
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)
	err = json.Unmarshal(body, &response)
	tests.AssertNoError(t, err)

	// Verify response
	tests.AssertEqual(t, validUUID, response["id"])
	tests.AssertEqual(t, true, response["valid"])
}

func TestController_UUIDValidation_InvalidUUID(t *testing.T) {
	// Setup environment and test helper
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Test handling of invalid UUID
	app := fiber.New()

	app.Get("/test/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Validate UUID
		if _, err := uuid.Parse(id); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid UUID"})
		}

		return c.JSON(fiber.Map{"id": id, "valid": true})
	})

	// Create request with invalid UUID
	req := httptest.NewRequest("GET", "/test/invalid-uuid", nil)

	// Execute request
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 400, resp.StatusCode)

	// Parse error response
	var response map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)
	err = json.Unmarshal(body, &response)
	tests.AssertNoError(t, err)

	// Verify error response
	tests.AssertEqual(t, "Invalid UUID", response["error"])
}

func TestController_QueryParameters_Success(t *testing.T) {
	// Setup environment and test helper
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Test query parameter handling
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		restart := c.QueryBool("restart", false)
		override := c.QueryBool("override", false)
		format := c.Query("format", "json")

		return c.JSON(fiber.Map{
			"restart":  restart,
			"override": override,
			"format":   format,
		})
	})

	// Create request with query parameters
	req := httptest.NewRequest("GET", "/test?restart=true&override=false&format=xml", nil)

	// Execute request
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)
	err = json.Unmarshal(body, &response)
	tests.AssertNoError(t, err)

	// Verify response
	tests.AssertEqual(t, true, response["restart"])
	tests.AssertEqual(t, false, response["override"])
	tests.AssertEqual(t, "xml", response["format"])
}

func TestController_HTTPMethods_Success(t *testing.T) {
	// Setup environment and test helper
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Test different HTTP methods
	app := fiber.New()

	var getCalled, postCalled, putCalled, deleteCalled bool

	app.Get("/test", func(c *fiber.Ctx) error {
		getCalled = true
		return c.JSON(fiber.Map{"method": "GET"})
	})

	app.Post("/test", func(c *fiber.Ctx) error {
		postCalled = true
		return c.JSON(fiber.Map{"method": "POST"})
	})

	app.Put("/test", func(c *fiber.Ctx) error {
		putCalled = true
		return c.JSON(fiber.Map{"method": "PUT"})
	})

	app.Delete("/test", func(c *fiber.Ctx) error {
		deleteCalled = true
		return c.JSON(fiber.Map{"method": "DELETE"})
	})

	// Test GET
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 200, resp.StatusCode)
	tests.AssertEqual(t, true, getCalled)

	// Test POST
	req = httptest.NewRequest("POST", "/test", nil)
	resp, err = app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 200, resp.StatusCode)
	tests.AssertEqual(t, true, postCalled)

	// Test PUT
	req = httptest.NewRequest("PUT", "/test", nil)
	resp, err = app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 200, resp.StatusCode)
	tests.AssertEqual(t, true, putCalled)

	// Test DELETE
	req = httptest.NewRequest("DELETE", "/test", nil)
	resp, err = app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 200, resp.StatusCode)
	tests.AssertEqual(t, true, deleteCalled)
}

func TestController_ErrorHandling_StatusCodes(t *testing.T) {
	// Setup environment and test helper
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Test different error status codes
	app := fiber.New()

	app.Get("/400", func(c *fiber.Ctx) error {
		return c.Status(400).JSON(fiber.Map{"error": "Bad Request"})
	})

	app.Get("/401", func(c *fiber.Ctx) error {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	})

	app.Get("/403", func(c *fiber.Ctx) error {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	})

	app.Get("/404", func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{"error": "Not Found"})
	})

	app.Get("/500", func(c *fiber.Ctx) error {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
	})

	// Test different status codes
	testCases := []struct {
		path string
		code int
	}{
		{"/400", 400},
		{"/401", 401},
		{"/403", 403},
		{"/404", 404},
		{"/500", 500},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest("GET", tc.path, nil)
		resp, err := app.Test(req)
		tests.AssertNoError(t, err)
		tests.AssertEqual(t, tc.code, resp.StatusCode)
	}
}

func TestController_ConfigurationModel_JSONSerialization(t *testing.T) {
	// Setup environment and test helper
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Test Configuration model JSON serialization
	app := fiber.New()

	app.Get("/config", func(c *fiber.Ctx) error {
		config := &model.Configuration{
			UdpPort:         model.IntString(9231),
			TcpPort:         model.IntString(9232),
			MaxConnections:  model.IntString(30),
			LanDiscovery:    model.IntString(1),
			RegisterToLobby: model.IntString(1),
			ConfigVersion:   model.IntString(1),
		}
		return c.JSON(config)
	})

	// Create request
	req := httptest.NewRequest("GET", "/config", nil)

	// Execute request
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 200, resp.StatusCode)

	// Parse response
	var response model.Configuration
	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)
	err = json.Unmarshal(body, &response)
	tests.AssertNoError(t, err)

	// Verify response
	tests.AssertEqual(t, model.IntString(9231), response.UdpPort)
	tests.AssertEqual(t, model.IntString(9232), response.TcpPort)
	tests.AssertEqual(t, model.IntString(30), response.MaxConnections)
	tests.AssertEqual(t, model.IntString(1), response.LanDiscovery)
	tests.AssertEqual(t, model.IntString(1), response.RegisterToLobby)
	tests.AssertEqual(t, model.IntString(1), response.ConfigVersion)
}

func TestController_UserModel_JSONSerialization(t *testing.T) {
	// Setup environment and test helper
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Test User model JSON serialization (password should be hidden)
	app := fiber.New()

	app.Get("/user", func(c *fiber.Ctx) error {
		user := &model.User{
			ID:       uuid.New(),
			Username: "testuser",
			Password: "secret-password", // Should not appear in JSON
			RoleID:   uuid.New(),
		}
		return c.JSON(user)
	})

	// Create request
	req := httptest.NewRequest("GET", "/user", nil)

	// Execute request
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 200, resp.StatusCode)

	// Parse response as raw JSON to check password is excluded
	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)

	// Verify password field is not in JSON
	if bytes.Contains(body, []byte("password")) || bytes.Contains(body, []byte("secret-password")) {
		t.Fatal("Password should not be included in JSON response")
	}

	// Verify other fields are present
	if !bytes.Contains(body, []byte("username")) || !bytes.Contains(body, []byte("testuser")) {
		t.Fatal("Username should be included in JSON response")
	}
}

func TestController_MiddlewareChaining_Success(t *testing.T) {
	// Setup environment and test helper
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Test middleware chaining
	app := fiber.New()

	var middleware1Called, middleware2Called, handlerCalled bool

	// Middleware 1
	middleware1 := func(c *fiber.Ctx) error {
		middleware1Called = true
		c.Locals("middleware1", "executed")
		return c.Next()
	}

	// Middleware 2
	middleware2 := func(c *fiber.Ctx) error {
		middleware2Called = true
		c.Locals("middleware2", "executed")
		return c.Next()
	}

	// Handler
	handler := func(c *fiber.Ctx) error {
		handlerCalled = true
		return c.JSON(fiber.Map{
			"middleware1": c.Locals("middleware1"),
			"middleware2": c.Locals("middleware2"),
			"handler":     "executed",
		})
	}

	app.Get("/test", middleware1, middleware2, handler)

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)

	// Execute request
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 200, resp.StatusCode)

	// Verify all were called
	tests.AssertEqual(t, true, middleware1Called)
	tests.AssertEqual(t, true, middleware2Called)
	tests.AssertEqual(t, true, handlerCalled)

	// Parse response
	var response map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)
	err = json.Unmarshal(body, &response)
	tests.AssertNoError(t, err)

	// Verify middleware values were passed
	tests.AssertEqual(t, "executed", response["middleware1"])
	tests.AssertEqual(t, "executed", response["middleware2"])
	tests.AssertEqual(t, "executed", response["handler"])
}
