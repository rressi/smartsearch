package smartsearch

import (
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"io"
    "unicode"
	"golang.org/x/text/language"
	"golang.org/x/text/cases"
	"golang.org/x/text/runes"
)

func ReadNormalized(r io.Reader) io.Reader {

	removeMn := func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}

	replaceInvalidChars := runes.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		} else {
			return rune(' ')
		}
	})

	spaceCount := 0
	removeRepeatedSpaces := func(r rune) bool {
		if r == rune(' ') {
			spaceCount += 1
			return spaceCount > 1
		} else {
			spaceCount = 0
			return false
		}
	}

	t := transform.Chain(
		norm.NFD,
		transform.RemoveFunc(removeMn),
		norm.NFC,
		cases.Lower(language.English),
		replaceInvalidChars,
		transform.RemoveFunc(removeRepeatedSpaces))

	return transform.NewReader(r, t)
}
