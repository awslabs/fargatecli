package cmd

import "strings"

// Humanize takes strings intended for machines and prettifies them for humans.
func Humanize(s string) string {
	s = strings.Replace(s, "_", " ", -1)
	s = strings.ToLower(s)
	s = strings.Title(s)

	return s
}

// Map applies a func to all members of a slice of strings and returns a new slice of the results.
func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))

	for i, v := range vs {
		vsm[i] = f(v)
	}

	return vsm
}
