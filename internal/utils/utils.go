package utils

import (
	"bytes"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

func StringToInteger(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func IntegerToString(i int) string {
	s := strconv.Itoa(i)
	return s
}

func ResetRequestBody(c *gin.Context, b []byte) {
	c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
}

func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
