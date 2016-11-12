package smartsearch

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
	"unicode"
	"unicode/utf8"
)

type Normalizer interface {
	Apply(src string) (result string)
}

type normalizerImpl struct {
	normalizationMap []rune
	spaceCount       int
}

const MAX_MAP = (2 << 16)

func NewNormalizer() Normalizer {

	normalizer := new(normalizerImpl)

	// Uses the transformer to generate a normalization map:
	normalizer.normalizationMap = make([]rune, MAX_MAP)
	toLowerCase := cases.Lower(language.English)
	for r := rune(0); r < rune(MAX_MAP); r++ {

		// We normalize only letters and digits, everything else is considered
		// as a separator:
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			normalizer.normalizationMap[r] = 0
			continue
		}

		// Takes the rune as UTF-8:
		var src [16]byte
		nSrc := utf8.EncodeRune(src[:], r)

		// To lower case:
		var dst [16]byte
		nDst, _, err := toLowerCase.Transform(dst[:], src[:nSrc], true)
		if nDst == 0 || err != nil {
			normalizer.normalizationMap[r] = 0
			panic(fmt.Sprintf("Cannot lower-case rune '%c' (%v): %v\n", r, r,
				err))
		}

		// Unicode normalization:
		var dst2 [16]byte
		nDst2, _, err := norm.NFD.Transform(dst2[:], dst[:nDst], true)
		if nDst2 == 0 || err != nil {
			normalizer.normalizationMap[r] = 0
			panic(fmt.Sprintf("Cannot normalize rune '%c' (%v): %v\n", r, r,
				err))
		}

		// From UTF-8:
		var nr rune
		nr, _ = utf8.DecodeRune(dst2[:nDst2])
		if nr == 0 {
			panic(fmt.Sprintf("Cannot decode rune '%c' (%v) fron bytes %v\n",
				r, r, dst2[:nDst2]))
		}

		/*
			if r == nr {
				fmt.Printf("[%c] %v\n", r, r)
			} else {
				fmt.Printf("[%c] %v -> [%c] %v\n", r, r, nr, nr)
			}
		*/

		normalizer.normalizationMap[r] = nr
	}

	return normalizer
}

func (n *normalizerImpl) Apply(src string) (result string) {

	nSrc := len(src)
	if nSrc == 0 {
		return
	}

	dst := make([]byte, len(src))
	nDst := 0
	nSeparators := 1
	for _, r := range src {
		if r < MAX_MAP {
			r = n.normalizationMap[r]
		}
		if r == 0 {
			r = ' '
			nSeparators++
		} else {
			nSeparators = 0
		}
		if nSeparators <= 1 {
			nDst += utf8.EncodeRune(dst[nDst:], r)
		}
	}

	result = string(dst[:nDst])
	return
}
