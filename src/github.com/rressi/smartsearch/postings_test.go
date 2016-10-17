package smartsearch

import (
	"reflect"
	"testing"
)

func TestPostings_CopyPostings(t *testing.T) {
	source := []int{10, 2, 5, 3, 5}
	expected_result := source

	result := CopyPostings(source)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}
}

func TestPostings_DedupPostings(t *testing.T) {
	source := []int{10, 2, 5, 3, 5}
	expected_result := []int{2, 3, 5, 10}

	result := SortDedupPostings(source)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}
}

func TestPostings_MergePostings(t *testing.T) {
	sourceA := []int{2, 3, 5, 10}
	sourceB := []int{1, 2, 5, 11}
	expected_result := []int{2, 5}

	result := MergePostings(sourceA, sourceB)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}
}
