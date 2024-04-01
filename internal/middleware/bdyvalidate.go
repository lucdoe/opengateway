package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"reflect"
)

type JSONValidator interface {
	Process(next http.Handler) http.Handler
}

type concreteJSONValidator struct {
	ExpectedFields map[string]string
}

func NewJSONValidator(expectedFields map[string]string) JSONValidator {
	return &concreteJSONValidator{ExpectedFields: expectedFields}
}

func (v *concreteJSONValidator) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := v.validateRequestBody(r); err != nil {
			http.Error(w, fmt.Sprintf("Request validation error: %v", err), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (v *concreteJSONValidator) validateRequestBody(r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("could not read request body: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var requestBody map[string]interface{}
	if err := json.Unmarshal(body, &requestBody); err != nil {
		return fmt.Errorf("could not unmarshal JSON: %v", err)
	}

	for field, fieldType := range v.ExpectedFields {
		if value, ok := requestBody[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		} else if !validateType(value, fieldType) {
			return fmt.Errorf("field %s is not of type %s", field, fieldType)
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
