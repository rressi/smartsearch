package smartsearch

import (
	"testing"
)

func TestNormalizer_Base(t *testing.T) {

	normalizer := NewNormalizer()

	var query, normalized, expectedNormalized string
	query = "This ìs ä fÄncy,  string"
	expectedNormalized = "this is a fancy string"
	normalized = normalizer.Apply(query)
	if normalized != expectedNormalized {
		t.Errorf("Unexpected result: '%v' and not '%v'", normalized,
			expectedNormalized)
	}
}
