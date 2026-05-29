package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckExportType(t *testing.T) {
	tests := []struct {
		path string
		want ExportType
	}{
		{"results.json", JSON},
		{"results.JSON", JSON},
		{"results.csv", CSV},
		{"results.txt", NotSupported},
	}

	for _, tt := range tests {
		if got := CheckExportType(tt.path); got != tt.want {
			t.Fatalf("CheckExportType(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestNewWriterUnsupported(t *testing.T) {
	_, err := NewWriter("out.txt", nil)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestJSONWriterRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	items := []Item{
		{Title: "Twitter", URL: "https://twitter.com/alice", Found: true},
	}

	writer, err := NewWriter(path, items)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	if err := writer.Write(); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	var got []Item
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if len(got) != 1 || got[0].Title != "Twitter" || !got[0].Found {
		t.Fatalf("unexpected data: %+v", got)
	}
}

func TestCSVWriterRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")
	items := []Item{
		{Title: "GitHub", URL: "https://github.com/alice", Found: false},
	}

	writer, err := NewWriter(path, items)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	if err := writer.Write(); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "GitHub") || !strings.Contains(content, "false") {
		t.Fatalf("unexpected csv content: %q", content)
	}
}
