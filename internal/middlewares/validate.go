package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
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
			c.Abort()
		}

		if err := checkFields(requestBody, expectedBody.Fields); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
		}
	}
}

func readAndParseJSON(c *gin.Context) (map[string]interface{}, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read request body")
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var requestBody map[string]interface{}
	if err := json.Unmarshal([]byte(body), &requestBody); err != nil {
		return nil, fmt.Errorf("invalid JSON")
	}

	return requestBody, nil
}

func checkFields(requestBody map[string]interface{}, expectedFields map[string]interface{}) error {
	for key, expectedValue := range expectedFields {
		if value, exists := requestBody[key]; exists {
			if err := handleField(value, expectedValue, key); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unvalid body")
		}
	}
	return nil
}

func handleField(value interface{}, expectedValue interface{}, key string) error {
	if reflect.TypeOf(expectedValue).Kind() == reflect.Map {
		return handleNestedField(value, expectedValue, key)
	} else {
		return handleSimpleField(value, expectedValue, key)
	}
}

func handleNestedField(value interface{}, expectedValue interface{}, key string) error {
	if reflect.TypeOf(value).Kind() != reflect.Map {
		return fmt.Errorf("invalid type for field '%s', expected object", key)
	}
	nestedExpectedFields, _ := expectedValue.(map[string]interface{})
	nestedFields, _ := value.(map[string]interface{})
	return checkFields(nestedFields, nestedExpectedFields)
}

func handleSimpleField(value interface{}, expectedValue interface{}, key string) error {
	expectedType, ok := expectedValue.(string)
	if !ok || !validateType(value, expectedType) {
		return fmt.Errorf("invalid type for field '%s', expected %s", key, expectedType)
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
