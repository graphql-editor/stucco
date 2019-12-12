package utils

import (
	"strings"
)

// FieldName returns resolver name for field
func FieldName(parent, field string) string {
	return strings.Join([]string{parent, field}, ".")
}
