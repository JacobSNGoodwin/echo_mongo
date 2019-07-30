package util

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

// GetUID utility extracts Username from jwt via ECHO middleware
func GetUID(c echo.Context) string {
	return c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["userId"].(string)
}

// GetUserName utility extracts Username from jwt via ECHO middleware
func GetUserName(c echo.Context) string {
	return c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["userName"].(string)
}

// ContainsImage assures that the MIME-Type of an uploaded file is a supported image format
func ContainsImage(s []string) bool {
	for _, a := range s {
		if a == "image/jpeg" || a == "image/gif" || a == "image/png" || a == "image/svg+xml" || a == "image/webp" {
			return true
		}
	}
	return false
}
