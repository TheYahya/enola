package exporter

import (
	"encoding/json"
	"fmt"
)

type JsonWriter struct {
	OutputPath string
	Items      []Item
}

func (writer JsonWriter) Write() {
	file, err := OpenOrCreateFile(writer.OutputPath)
	if err != nil {
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(writer.Items); err != nil {
		fmt.Println("Error while saving items to JSON file", err)
	}
}
