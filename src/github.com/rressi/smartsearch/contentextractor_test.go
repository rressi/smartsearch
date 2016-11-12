package smartsearch

import (
	"testing"
)

func TestContentExtractor_Base(t *testing.T) {
	source := "{\"id\":10, " +
		"\"title\":\"some title\", " +
		"\"content\":\"some content\", " +
		"\"extra\":[1, 2, 3]}"
	expected_content := "some title some content"

	jsonExtractor := MakeJsonExtractor("id", []string{"title",
		"content", "extra"})
	id, content, err := jsonExtractor([]byte(source))
	if err != nil {
		t.Errorf("Failed: %v", err)
	} else if id != 10 {
		t.Errorf("Invalid id: %v", id)
	} else if content != expected_content {
		t.Errorf("Unexpected content: '%v'", content)
	}
}
