package main

import "strings"

var allowedLanguages = map[string]struct{}{
	"c":      {},
	"go":     {},
	"python": {},
	"ruby":   {},
}

func normalizeLanguage(lang string) (string, bool) {
	n := strings.ToLower(strings.TrimSpace(lang))
	_, ok := allowedLanguages[n]
	if !ok {
		return "", false
	}
	return n, true
}
