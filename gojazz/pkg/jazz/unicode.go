package jazz

import "unicode"

func isAlphaNumeric(r rune) bool {
	return isDigit(r) || isAlpha((r))
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func isAlpha(r rune) bool {
	return unicode.IsLetter(r)
}
