package main

import (
	"regexp"
	"strings"
)

func FindFromRegExp(s string, t string) string {
	r := ""
	re := regexp.MustCompile(t)
	match := re.FindStringSubmatch(s)
	if match != nil && len(match) > 1 {
		r = match[1]
	}
	return r
}

func ReplaceBadSymbols(s string) string {
	x := strings.ReplaceAll(s, "\u0000", "")
	x = strings.ReplaceAll(x, "\u0026", "")
	return x
}
