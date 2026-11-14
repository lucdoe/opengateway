// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// variableName is what may stand between ${ and } (or :-).
var variableName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// fillPlaceholders replaces ${NAME} in every value of doc with the
// environment variable NAME, and ${NAME:-fallback} with fallback when NAME is
// unset or empty. $${ stands for a literal ${. Keys are left alone.
//
// It works on the parsed document, not the text, so a value can never change
// the structure of the config, and errors keep the line they come from.
// Every problem is reported in one error.
func fillPlaceholders(doc *yaml.Node) error {
	var problems []string
	walkValues(doc, func(n *yaml.Node) {
		if !strings.Contains(n.Value, "${") {
			return
		}
		filled, err := expand(n.Value)
		if err != nil {
			problems = append(problems, fmt.Sprintf("line %d: %v", n.Line, err))
			return
		}
		n.Value = filled
		if n.Style == 0 {
			// Unquoted, so type it by what it now says, as if it had been
			// written there: Port: ${PORT} becomes a number.
			n.Tag = ""
		}
	})
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("the config has placeholders that cannot be filled:\n- %s", strings.Join(problems, "\n- "))
}

// walkValues calls fill for every scalar that is a value, not a key.
func walkValues(n *yaml.Node, fill func(*yaml.Node)) {
	switch n.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, child := range n.Content {
			walkValues(child, fill)
		}
	case yaml.MappingNode:
		for i := 1; i < len(n.Content); i += 2 {
			walkValues(n.Content[i], fill)
		}
	case yaml.ScalarNode:
		fill(n)
	}
}

// expand fills the placeholders of one value. It stops at the first problem.
func expand(value string) (string, error) {
	var out strings.Builder
	for {
		start := strings.Index(value, "${")
		if start < 0 {
			out.WriteString(value)
			return out.String(), nil
		}
		if start > 0 && value[start-1] == '$' { // $${ is a literal ${
			out.WriteString(value[:start-1] + "${")
			value = value[start+2:]
			continue
		}
		out.WriteString(value[:start])

		length := strings.IndexByte(value[start:], '}')
		if length < 0 {
			return "", fmt.Errorf("%s has no closing }", value[start:])
		}
		placeholder := value[start : start+length+1]
		name, fallback, hasFallback := strings.Cut(placeholder[2:length], ":-")
		if !variableName.MatchString(name) {
			return "", fmt.Errorf("%s: %q is not a variable name (letters, digits and _, not starting with a digit)", placeholder, name)
		}
		filled := os.Getenv(name)
		if filled == "" {
			if !hasFallback {
				return "", fmt.Errorf("%s: %s is not set; set it, or give a fallback with ${%s:-fallback}", placeholder, name, name)
			}
			filled = fallback
		}
		out.WriteString(filled)
		value = value[start+length+1:]
	}
}
