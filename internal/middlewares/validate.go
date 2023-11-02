package middlewares

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidateJSONFields(allowedJSON []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read request body"})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// reverse HTML escape
		decodedBody := html.UnescapeString(string(body))
		var requestBody map[string]interface{}

		err = json.Unmarshal([]byte(decodedBody), &requestBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		for key := range requestBody {
			if !Contains(allowedJSON, key) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

		}

	}
}

func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
