package utils

import (
	"strings"
)

func FieldName(parent, field string) string {
	return strings.Join([]string{parent, ".", field}, "")
}
