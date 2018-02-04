package cmd

import "strings"

// Humanize takes strings intended for machines and prettifies them for humans.
func Humanize(s string) string {
	s = strings.Replace(s, "_", " ", -1)
	s = strings.ToLower(s)
	s = strings.Title(s)

	return s
}
