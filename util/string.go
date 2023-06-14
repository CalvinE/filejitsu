package util

func PadLeft(s string, padChar string, desiredLength int) string {
	initialLength := len(s)
	padCharLength := len(padChar)
	if initialLength >= desiredLength {
		return s
	}
	for i := initialLength; i < desiredLength; i += padCharLength {
		s = padChar + s
	}
	return s
}

func PadRight(s string, padChar string, desiredLength int) string {
	initialLength := len(s)
	padCharLength := len(padChar)
	if initialLength >= desiredLength {
		return s
	}
	for i := initialLength; i < desiredLength; i += padCharLength {
		s = s + padChar
	}
	return s
}
