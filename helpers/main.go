package helpers

import "strings"

func NormalizeName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	return strings.ToLower(name)
}
