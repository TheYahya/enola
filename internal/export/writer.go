package export

import (
	"fmt"
	"strings"
)

// ExportType identifies a supported export format.
type ExportType string

const (
	JSON         ExportType = "json"
	CSV          ExportType = "csv"
	NotSupported ExportType = "notsupported"
)

// Item is a single export row.
type Item struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Found bool   `json:"found"`
}

// Writer persists export items.
type Writer interface {
	Write() error
}

type factory func(path string, items []Item) Writer

var writers = map[ExportType]factory{
	JSON: func(path string, items []Item) Writer {
		return jsonWriter{outputPath: path, items: items}
	},
	CSV: func(path string, items []Item) Writer {
		return csvWriter{outputPath: path, items: items}
	},
}

// CheckExportType returns the export type for a file path.
func CheckExportType(filename string) ExportType {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, string(JSON)) {
		return JSON
	}
	if strings.HasSuffix(lower, string(CSV)) {
		return CSV
	}
	return NotSupported
}

// NewWriter returns a Writer for the given path and items.
func NewWriter(path string, items []Item) (Writer, error) {
	exportType := CheckExportType(path)
	f, ok := writers[exportType]
	if !ok {
		return nil, fmt.Errorf("unsupported export format: %s", path)
	}
	return f(path, items), nil
}
