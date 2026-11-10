package files

import (
	"slices"
	"strings"
)

const AllowedSpecialCharacters = "-_. "

var ReservedNames = []string{"con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9"}

func IsValidFileName(filename string) bool {
	if len(filename) == 0 || len(filename) > 255 {
		return false
	}

	if filename == "." || filename == ".." {
		return false
	}

	if strings.HasPrefix(filename, "-") {
		return false
	}

	if strings.HasSuffix(filename, " ") || strings.HasSuffix(filename, ".") {
		return false
	}

	filename = strings.ToLower(filename)

	if strings.ContainsFunc(filename, func(r rune) bool {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || strings.ContainsRune(AllowedSpecialCharacters, r) {
			return false
		}
		return true
	}) {
		return false
	}

	indexOfFirstDot := strings.Index(filename, ".")
	basename := filename
	if indexOfFirstDot != -1 {
		basename = filename[:indexOfFirstDot]
	}
	if slices.Contains(ReservedNames, basename) {
		return false
	}

	return true
}
