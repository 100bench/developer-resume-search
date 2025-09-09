package utils

import (
	"html/template"
	"strings"
)

// Pluralize returns the plural form of a word if count is not 1.
func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

// SliceString slices a string to a given length and adds an ellipsis if truncated.
func SliceString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// Linebreaksbr replaces newline characters with <br> tags.
func Linebreaksbr(text string) template.HTML {
	replaced := strings.ReplaceAll(text, "\n", "<br>")
	return template.HTML(replaced)
}
