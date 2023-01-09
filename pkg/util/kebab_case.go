package util

import (
	"regexp"
	"strings"
)

func ToKebabCase(str string) string {
	return strings.ToLower(regexp.MustCompile(`([a-z])([A-Z])`).ReplaceAllString(str, "$1-$2"))
}
