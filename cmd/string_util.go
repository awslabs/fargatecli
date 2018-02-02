package cmd

import "strings"

func Humanize(s string) string {
	s = strings.Replace(s, "_", " ", -1)
	s = strings.ToLower(s)
	s = strings.Title(s)

	return s
}
