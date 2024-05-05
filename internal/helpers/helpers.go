package helpers

import "github.com/gin-gonic/gin"

// handleError simplifies error response handling
func HandleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"code": statusCode, "status": "error", "data": err.Error()})
}

// handleSuccess simplifies success response handling
func HandleSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{"code": statusCode, "status": "success", "data": data})
}

func SetHeaders(c *gin.Context, headers map[string]string) {
	for header, value := range headers {
		c.Header(header, value)
	}
}
