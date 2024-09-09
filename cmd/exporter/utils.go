package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func OpenOrCreateFile(filename string) (*os.File, error) {
	dir := filepath.Dir(filename)
	// 0755 permissions: the owner can read, write, and execute, everyone else can only read and execute.
	err := os.MkdirAll(dir, 0755)

	if err != nil {
		fmt.Println("Error creating directory:", err)
		return nil, err
	}

	// Open the file with read and write permissions
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return file, nil
}

func CheckExportType(filename string) ExportType {
	lowerFileName := strings.ToLower(filename)
	if strings.HasSuffix(strings.ToLower(lowerFileName), string(JSON)) {
		return JSON
	} else if strings.HasSuffix(strings.ToLower(lowerFileName), string(CSV)) {
		return CSV
	}

	return NOTSUPPORTED
}
