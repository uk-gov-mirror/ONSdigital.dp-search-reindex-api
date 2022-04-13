package api

import "fmt"

// GetValueType gets the type of a value and returns the type which is understandable for an API user
func GetValueType(value interface{}) string {
	valueType := fmt.Sprintf("%T", value)

	switch valueType {
	case "bool":
		return "boolean"
	case "float32", "float64", "int", "int8", "int16", "int32", "int64":
		return "integer"
	default:
		return valueType
	}
}
