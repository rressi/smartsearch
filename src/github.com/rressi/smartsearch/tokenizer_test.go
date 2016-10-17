package smartsearch

import (
	"reflect"
	"testing"
)

func TestTokenizer_Base(t *testing.T) {

	tokens := Tokenize("YES!-This ìs ä fÄncy, is a string")
	expected_tokens := []string{"a", "fancy", "is", "string", "this", "yes"}
	if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v", tokens)
	}
}

func TestTokenizer_Empty(t *testing.T) {

	var tokens []string
	var expected_tokens []string

	tokens = Tokenize("")
	if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v,%d expected_tokens=%v,%d",
			tokens, len(tokens),
			expected_tokens, len(expected_tokens))
	}

	tokens = Tokenize(" ")
	if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v,%d expected_tokens=%v,%d",
			tokens, len(tokens),
			expected_tokens, len(expected_tokens))
	}

	tokens = Tokenize("_/@--")
	if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v,%d expected_tokens=%v,%d",
			tokens, len(tokens),
			expected_tokens, len(expected_tokens))
	}
}
