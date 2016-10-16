package smartsearch

import (
	"bytes"
	"testing"
)

func TestTrieBuilder_Empty(t *testing.T) {

	builder := NewTrieBuilder()

	buf := new(bytes.Buffer)
	err := builder.Dump(buf)
	if err != nil {
		t.Errorf("Error while duming: %v", err)
	}

	if bytes.Compare(buf.Bytes(), []byte{0, 0}) != 0 {
		t.Errorf("Unexpected serialization: %v", buf.Bytes()[:10])
	}
}

func TestTrieBuilder_VoidTerm(t *testing.T) {

	builder := NewTrieBuilder()
	builder.Add(1, "")
	builder.Add(2, "")
	builder.Add(1, "")
	builder.Add(2, "")

	buf := new(bytes.Buffer)
	err := builder.Dump(buf)
	if err != nil {
		t.Errorf("Error while duming: %v", err)
	}

	if bytes.Compare(buf.Bytes(), []byte{2, 0, 2, 1, 1}) != 0 {
		t.Errorf("Unexpected serialization: %v", buf.Bytes()[:10])
	}
}

func TestTrieBuilder_Base(t *testing.T) {

	builder := NewTrieBuilder()
	builder.Add(1, "A")
	builder.Add(2, "A")
	builder.Add(1, "B")
	builder.Add(2, "B")

	buf := new(bytes.Buffer)
	err := builder.Dump(buf)
	if err != nil {
		t.Errorf("Error while duming: %v", err)
	}

	expected := []byte{0, 2, 4, 65, 5, 1, 5, 2, 0, 2, 1, 1, 2, 0, 2, 1, 1}
	if bytes.Compare(buf.Bytes(), expected) != 0 {
		t.Errorf("Unexpected serialization: %v", buf.Bytes())
	}
}

func TestTrieBuilder_Term(t *testing.T) {

	builder := NewTrieBuilder()
	builder.Add(1, "ABC")
	builder.Add(2, "BCA")
	builder.Add(3, "CAB")

	buf := new(bytes.Buffer)
	err := builder.Dump(buf)
	if err != nil {
		t.Errorf("Error while duming: %v", err)
	}

	expected := []byte{0, 3, 6, 65, 14, 1, 14, 1, 14, 0, 1, 2, 66, 9, 0, 1,
		2, 67, 4, 1, 0, 1, 1, 0, 1, 2, 67, 9, 0, 1, 2, 65, 4, 1, 0, 1, 2, 0,
		1, 2, 65, 9, 0, 1, 2, 66, 4, 1, 0, 1, 3}
	if bytes.Compare(buf.Bytes(), expected) != 0 {
		t.Errorf("Unexpected serialization: %v,%d", buf.Bytes(), buf.Len())
	}
}
