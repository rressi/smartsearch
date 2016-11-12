package smartsearch

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"unicode"
)

type Normalizer interface {
	Apply(src string) (result string, err error)
}

type normalizerImpl struct {
	transformer transform.Transformer
	spaceCount  int
}

func NewNormalizer() Normalizer {

	normalizer := new(normalizerImpl)
	normalizer.spaceCount = 1

	removeRepeatedSpaces := func(r rune) bool {
		if r == rune(' ') {
			normalizer.spaceCount += 1
			return normalizer.spaceCount > 1
		} else {
			normalizer.spaceCount = 0
			return false
		}
	}

	normalizer.transformer = transform.Chain(
		norm.NFD,
		transform.RemoveFunc(removeMn),
		norm.NFC,
		cases.Lower(language.English),
		runes.Map(replaceInvalidChars),
		transform.RemoveFunc(removeRepeatedSpaces))

	return normalizer
}

func (n *normalizerImpl) Apply(src string) (result string, err error) {

	dst := make([]byte, len(src))
	nDst, nSrc, err := n.transformer.Transform(dst, []byte(src), true)
	if err != nil {
		err = fmt.Errorf("Normalizer.Apply: %v", err)
		return
	} else if nSrc != len(src) {
		err = fmt.Errorf(
			"Normalizer.Apply: %v charcters processed instead of %v", nSrc,
			len(src))
		return
	}

	result = string(dst[:nDst])
	return
}

func removeMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}

func replaceInvalidChars(r rune) rune {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return r
	} else {
		return rune(' ')
	}
}
