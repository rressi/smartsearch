package smartsearch

import (
	"sort"
)

// It just clones one sequence of postings into new one.
func CopyPostings(src []int) (postings []int) {
	for _, posting := range src {
		postings = append(postings, posting)
	}
	return
}

// It takes a sequence of postings and generates a new one that is obtained
// after sorting and deduplicating them.
func SortDedupPostings(src []int) (postings []int) {

	if len(src) == 0 {
		return
	}

	postings = CopyPostings(src)

	if len(postings) > 1 {
		sort.Ints(postings)
		var i int
		for j := 1; j < len(postings); j++ {
			if postings[i] != postings[j] {
				i++
				postings[i] = postings[j]
			}
		}
		postings = postings[:i+1]
	}

	return
}

// It takes 2 sorted and deduplicated sequences of postings and generates a new
// sorted and deduplicated sequence that contains the intersection of
// the postings.
func IntersectPostings(srcA []int, srcB []int) (postings []int) {

	nA := len(srcA)
	nB := len(srcB)
	if nA == 0 || nB == 0 {
		return // No results!
	}

	var iA, iB int
	for iA < nA && iB < nB {
		a := srcA[iA]
		b := srcB[iB]
		if a < b {
			iA++
		} else if a > b {
			iB++
		} else {
			postings = append(postings, a)
			iA++
			iB++
		}
	}

	return
}

// It takes 2 sorted and deduplicated sequences of postings and generates a new
// sorted and deduplicated sequence that contains the union of
// the postings.
func UnitePostings(srcA []int, srcB []int) (postings []int) {

	nA := len(srcA)
	nB := len(srcB)

	// Handles corner cases:
	if nA == 0 {
		postings = srcB
		return
	} else if nA == 0 {
		postings = srcA
		return
	}

	// Fuses the 2 sequences:
	var iA, iB int
	for iA < nA && iB < nB {
		a := srcA[iA]
		b := srcB[iB]
		if a < b {
			postings = append(postings, a)
			iA++
		} else if a > b {
			postings = append(postings, b)
			iB++
		} else {
			postings = append(postings, a)
			iA++
			iB++
		}
	}

	// Attaches the eventual tail:
	if iA < nA {
		postings = append(postings, srcA[iA:]...)
	} else if iB < nB {
		postings = append(postings, srcB[iB:]...)
	}

	return
}
