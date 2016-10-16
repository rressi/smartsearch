package smartsearch

import (
	"reflect"
	"testing"
)

func TestTokenizer_Base(t *testing.T) {

	tokens := Tokenize("YES!-This ìs ä fÄncy,  string")
	expected_tokens := []string{"a", "fancy", "is", "string", "this", "yes"}
	if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v", tokens)
	}
}
