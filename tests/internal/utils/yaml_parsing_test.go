package utils

import (
	"reflect"
	"testing"

	"github.com/lucdoe/capstone_gateway/internal/utils"
)

func TestYAMLParsingUnmarshal(t *testing.T) {
	yamlInput := []byte(`
name: John Doe
age: 30
`)

	expected := struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}{
		Name: "John Doe",
		Age:  30,
	}

	parserInstance := utils.YAMLParsing{}

	var actual struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}

	err := parserInstance.Unmarshal(yamlInput, &actual)
	if err != nil {
		t.Fatalf("Unmarshal returned an error: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Unmarshal = %v, want %v", actual, expected)
	}
}
