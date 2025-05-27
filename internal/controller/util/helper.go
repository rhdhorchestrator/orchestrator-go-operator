// util/helper.go
package util

func MakePointer[T any](t T) *T {
	return &t
}

// ValidateMap Helper to ensure nested map[string]interface{} structure
func ValidateMap(m map[string]interface{}, key string) map[string]interface{} {
	if val, ok := m[key]; ok {
		if cast, ok := val.(map[string]interface{}); ok {
			return cast
		}
	}
	newMap := make(map[string]interface{})
	m[key] = newMap
	return newMap
}

// ValidateSlice Helper to ensure nested []interface{} structure
func ValidateSlice(m map[string]interface{}, key string) []interface{} {
	if val, ok := m[key]; ok {
		if cast, ok := val.([]interface{}); ok {
			return cast
		}
	}
	return []interface{}{}
}
