package handlers

import (
	"regexp"
)

var regexpSymbols *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isValidStr(str string, isKey bool) bool {
	if !isKey && len(str) > 20 && len(str) < 6 {
		return false
	}

	return regexpSymbols.MatchString(str)
}
