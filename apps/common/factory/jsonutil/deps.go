// Package jsonutil provides JSON helpers for task dependency arrays.
package jsonutil

import "encoding/json"

// ParseDeps parses a JSON array string into a slice of strings.
// Returns nil for empty, "null", or "[]" input.
func ParseDeps(jsonStr string) []string {
	if jsonStr == "" || jsonStr == "null" || jsonStr == "[]" {
		return nil
	}
	var deps []string
	if err := json.Unmarshal([]byte(jsonStr), &deps); err != nil {
		return nil
	}
	return deps
}

// ToJSON marshals v to a JSON string. Returns "[]" on error.
func ToJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// ContainsDep checks if depsJSON (a JSON array) contains taskID.
func ContainsDep(depsJSON string, taskID string) bool {
	deps := ParseDeps(depsJSON)
	for _, d := range deps {
		if d == taskID {
			return true
		}
	}
	return false
}
