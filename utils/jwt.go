package utils

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func ErrorResponseUnauthorizedJwt(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"status":  http.StatusUnauthorized,
		"message": "Failed",
		"data":    nil,
		"error":   "[middleware][jwt] " + message,
	})
}

func JWTAuth() gin.HandlerFunc {
	secret := os.Getenv("JWT_SECRET")

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			ErrorResponseUnauthorizedJwt(c, "[middleware][jwt] Authorization header prefix is not provided")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ErrorResponseUnauthorizedJwt(c, "[middleware][jwt] Invalid or missing Bearer token")
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil {
			ErrorResponseUnauthorizedJwt(c, "[middleware][jwt] Invalid token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			ErrorResponseUnauthorizedJwt(c, "[middleware][jwt] Invalid token")
			c.Abort()
			return
		}

		c.Set("user_id", claims["sub"].(float64))
		c.Set("email", claims["email"].(string))
		c.Next()
	}
}
