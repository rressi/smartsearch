package smartsearch

import (
	"sort"
)

func CopyPostings(src []int) (postings []int) {
	for _, posting := range src {
		postings = append(postings, posting)
	}
	return
}

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

func MergePostings(srcA []int, srcB []int) (postings []int) {

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
