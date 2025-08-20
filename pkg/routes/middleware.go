package routes

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

// APIKeyMiddleware validates X-API-Key header against expected key
func APIKeyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		println(c.Request())
		apiKey := c.Request().Header.Get("X-API-Key")
		expected := os.Getenv("API_KEY")

		if apiKey == "" || apiKey != expected {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid or missing API key"})
		}

		return next(c)
	}
}
