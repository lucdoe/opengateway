// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Keys the gateway no longer reads, with what to do instead. The strict
// decoder would only report them as unknown.
var (
	retiredTopKeys = map[string]string{
		"Redis": "the top-level Redis section has moved: set Store.Type: redis and put its settings under Store.Redis",
	}
	retiredServiceKeys = map[string]string{
		"Protocol": "Protocol has been removed: the scheme of URL sets it",
	}
	retiredEndpointKeys = map[string]string{
		"QueryParams": "QueryParams has been removed: it was never applied to requests",
		"Body":        "Body has been removed: it was never applied to requests",
	}
	retiredAuthKeys = map[string]string{
		"ApplyAuth": `Auth.ApplyAuth has been removed: add "auth" to the endpoint's Plugins, or the service's for all its endpoints`,
		"Scope":     "Auth.Scope has moved: write it as the endpoint's Scope",
		"Method":    "Auth.Method has been removed: auth always expects a Bearer token",
		"Algorithm": "Auth.Algorithm has been removed: set Plugins.JWTConfig.SigningMethod",
		"SecretKey": "Auth.SecretKey has been removed: set Plugins.JWTConfig.SecretKey or JWT_SECRET",
	}
)

func rejectRetiredKeys(data []byte) error {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil || len(doc.Content) == 0 {
		return nil // the strict decode reports it
	}
	var problems []string
	check := func(node *yaml.Node, retired map[string]string) {
		for _, key := range mappingKeys(node) {
			if why, ok := retired[key.Value]; ok {
				problems = append(problems, fmt.Sprintf("line %d: %s", key.Line, why))
			}
		}
	}

	top := doc.Content[0]
	check(top, retiredTopKeys)
	services := child(top, "Services")
	for _, name := range mappingKeys(services) {
		service := child(services, name.Value)
		check(service, retiredServiceKeys)
		if endpoints := child(service, "Endpoints"); endpoints != nil && endpoints.Kind == yaml.SequenceNode {
			for _, endpoint := range endpoints.Content {
				check(endpoint, retiredEndpointKeys)
				check(child(endpoint, "Auth"), retiredAuthKeys)
			}
		}
	}

	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("the config uses settings that are no longer read:\n- %s", strings.Join(problems, "\n- "))
}

// mappingKeys returns the key nodes of a mapping node, or nil for any other
// node.
func mappingKeys(node *yaml.Node) []*yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	keys := make([]*yaml.Node, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		keys = append(keys, node.Content[i])
	}
	return keys
}

// child returns the value at key in a mapping node, or nil.
func child(node *yaml.Node, key string) *yaml.Node {
	for i, k := range mappingKeys(node) {
		if k.Value == key {
			return node.Content[2*i+1]
		}
	}
	return nil
}
