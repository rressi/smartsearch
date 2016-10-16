package smartsearch

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestNormalizer_Base(t *testing.T) {

	buf := bytes.NewBufferString("This ìs ä fÄncy,  string")
	reader := ReadNormalized(buf)
	if reader == nil {
		t.Error("Cannot create normalizer")
	}

	bytes_, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Errorf("Cannot read: %v", err)
	}

	normalized := string(bytes_)
	if normalized != "this is a fancy string" {
		t.Errorf("Unexpected result: %v", normalized)
	}
}
