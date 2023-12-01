package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal"
)

func ValidateJSONFields(expectedBody internal.BodyField) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestBody, err := readAndParseJSON(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := checkFieldsPresence(requestBody, expectedBody.Fields); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validateFieldsType(requestBody, expectedBody.Fields); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
}

func readAndParseJSON(c *gin.Context) (map[string]interface{}, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read request body")
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	decodedBody := html.UnescapeString(string(body))
	var requestBody map[string]interface{}
	if err := json.Unmarshal([]byte(decodedBody), &requestBody); err != nil {
		return nil, fmt.Errorf("invalid JSON")
	}

	return requestBody, nil
}

func checkFieldsPresence(requestBody map[string]interface{}, expectedFields map[string]interface{}) error {
	for key := range expectedFields {
		if _, exists := requestBody[key]; !exists {
			return fmt.Errorf("missing field '%s' in JSON body", key)
		}
	}
	return nil
}

func validateFieldsType(requestBody map[string]interface{}, expectedFields map[string]interface{}) error {
	for key, expectedTypeInterface := range expectedFields {
		expectedType, ok := expectedTypeInterface.(string)
		if !ok {
			return fmt.Errorf("internal configuration error")
		}
		if !validateType(requestBody[key], expectedType) {
			return fmt.Errorf("invalid type for field '%s', expected %s", key, expectedType)
		}
	}
	return nil
}

func validateType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "int":
		if _, ok := value.(float64); ok {
			floatValue := value.(float64)
			return floatValue == math.Floor(floatValue)
		}
		return false
	case "float":
		_, ok := value.(float64)
		return ok
	case "bool":
		_, ok := value.(bool)
		return ok
	case "array":
		val := reflect.ValueOf(value)
		return val.Kind() == reflect.Slice
	default:
		// For complex types
		return false
	}
}
