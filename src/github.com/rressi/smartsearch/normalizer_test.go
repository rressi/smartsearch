package smartsearch

import (
	"testing"
)

func TestNormalizer_Base(t *testing.T) {

	normalizer := NewNormalizer()
	normalized, err := normalizer.Apply("This ìs ä fÄncy,  string")
	if err != nil {
		t.Errorf("Normalization has failed: %v", err)
	} else if normalized != "this is a fancy string" {
		t.Errorf("Unexpected result: %v", normalized)
	}
}
