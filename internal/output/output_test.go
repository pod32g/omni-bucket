package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pod32g/omni-bucket/internal/output"
)

func TestTableRendersHeaderAndRows(t *testing.T) {
	var buf bytes.Buffer
	if err := output.Table(&buf, []string{"ID", "TITLE"}, [][]string{{"1", "hello"}}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "TITLE") || !strings.Contains(out, "hello") {
		t.Fatalf("missing content: %q", out)
	}
}

func TestJSONRendersIndented(t *testing.T) {
	var buf bytes.Buffer
	if err := output.JSON(&buf, map[string]int{"a": 1}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\"a\": 1") {
		t.Fatalf("got %q", buf.String())
	}
}
