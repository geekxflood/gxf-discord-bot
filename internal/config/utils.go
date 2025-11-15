package config

import (
	"regexp"
	"strconv"
	"strings"
)

// getNestedValue retrieves a value from a nested map using dot notation
// Supports both dot notation (foo.bar) and array indexing (foo[0].bar)
func getNestedValue(data map[string]interface{}, key string) (interface{}, bool) {
	parts := parseKey(key)
	current := interface{}(data)

	for _, part := range parts {
		// First handle map key
		if part.key != "" {
			m, ok := current.(map[string]interface{})
			if !ok {
				return nil, false
			}

			current, ok = m[part.key]
			if !ok {
				return nil, false
			}
		}

		// Then handle array index if present
		if idx, isArray := part.arrayIndex(); isArray {
			slice, ok := current.([]interface{})
			if !ok {
				return nil, false
			}
			if idx < 0 || idx >= len(slice) {
				return nil, false
			}
			current = slice[idx]
		}
	}

	return current, true
}

type keyPart struct {
	key   string
	index int
}

func (kp keyPart) arrayIndex() (int, bool) {
	if kp.index >= 0 {
		return kp.index, true
	}
	return 0, false
}

// parseKey parses a key string into parts
// Examples:
//   - "foo.bar" -> [{key: "foo"}, {key: "bar"}]
//   - "foo[0].bar" -> [{key: "foo", index: 0}, {key: "bar"}]
//   - "actions[1]" -> [{key: "actions", index: 1}]
func parseKey(key string) []keyPart {
	var parts []keyPart
	arrayIndexRegex := regexp.MustCompile(`^([^\[]+)\[(\d+)\]$`)

	segments := strings.Split(key, ".")
	for _, segment := range segments {
		if matches := arrayIndexRegex.FindStringSubmatch(segment); matches != nil {
			// Array access: "foo[0]"
			idx, _ := strconv.Atoi(matches[2])
			parts = append(parts, keyPart{
				key:   matches[1],
				index: idx,
			})
		} else {
			// Regular key: "foo"
			parts = append(parts, keyPart{
				key:   segment,
				index: -1,
			})
		}
	}

	return parts
}
