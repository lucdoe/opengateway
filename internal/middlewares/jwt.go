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
			tokenString := strings.Split(c.GetHeader("Authorization"), " ")[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				return []byte(secretKey), nil
			})

			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, please provide valid token"})
				c.Abort()
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				_ = claims
				c.Next()
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, invalid token"})
				c.Abort()
			}
		}
	}
	return func(c *gin.Context) {
		c.Next()
	}
}
