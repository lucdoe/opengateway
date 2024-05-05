package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func ValidateToken(shouldCheck bool, secretKey string) gin.HandlerFunc {
	if shouldCheck {
		return func(c *gin.Context) {
			if !validateAuthHeader(c) {
				return
			}

			tokenString := getTokenFromHeader(c)
			if !parseAndValidateToken(c, tokenString, secretKey) {
				return
			}

			c.Next()
		}
	}
	return func(c *gin.Context) {
		c.Next()
	}
}

func validateAuthHeader(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	splitToken := strings.Split(authHeader, " ")

	if len(splitToken) != 2 || splitToken[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized, please provide valid token in authorization header with format: Bearer {token}"})
		c.Abort()
		return false
	}
	return true
}

func getTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	splitToken := strings.Split(authHeader, " ")
	return splitToken[1]
}

func parseAndValidateToken(c *gin.Context, tokenString string, secretKey string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized, invalid token format"})
		c.Abort()
		return false
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized, invalid token"})
		c.Abort()
		return false
	}

	return true
}
