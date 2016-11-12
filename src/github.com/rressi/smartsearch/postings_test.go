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

func TestPostings_IntersectPostings(t *testing.T) {

	var sourceA, sourceB, expected_result, result []int

	sourceA = []int{2, 3, 5, 10}
	sourceB = []int{1, 2, 5, 11}
	expected_result = []int{2, 5}

	result = IntersectPostings(sourceA, sourceB)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}

	sourceA = []int{}
	sourceB = []int{1, 2, 5, 11}
	expected_result = nil

	result = IntersectPostings(sourceA, sourceB)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}

	sourceA = []int{2, 6, 7}
	sourceB = []int{}
	expected_result = nil

	result = IntersectPostings(sourceA, sourceB)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}

	sourceA = []int{2, 4, 6}
	sourceB = []int{1, 3, 5}
	expected_result = nil

	result = IntersectPostings(sourceA, sourceB)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}
}

func TestPostings_UnitePostings(t *testing.T) {

	var sourceA, sourceB, expected_result, result []int

	sourceA = []int{2, 3, 5, 10}
	sourceB = []int{1, 2, 5, 11}
	expected_result = []int{1, 2, 3, 5, 10, 11}

	result = UnitePostings(sourceA, sourceB)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}

	sourceA = []int{}
	sourceB = []int{1, 2, 5, 11}
	expected_result = []int{1, 2, 5, 11}

	result = UnitePostings(sourceA, sourceB)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}

	sourceA = []int{2, 6, 7}
	sourceB = []int{}
	expected_result = []int{2, 6, 7}

	result = UnitePostings(sourceA, sourceB)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}

	sourceA = []int{2, 4, 6, 7, 8}
	sourceB = []int{1, 3, 5}
	expected_result = []int{1, 2, 3, 4, 5, 6, 7, 8}

	result = UnitePostings(sourceA, sourceB)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}

	sourceA = []int{1, 3, 5}
	sourceB = []int{2, 4, 6, 7, 8}
	expected_result = []int{1, 2, 3, 4, 5, 6, 7, 8}

	result = UnitePostings(sourceA, sourceB)
	if !reflect.DeepEqual(result, expected_result) {
		t.Errorf("Unexpected result: %v", result)
	}
}
