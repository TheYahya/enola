package export

import (
	"encoding/json"
)

type jsonWriter struct {
	outputPath string
	items      []Item
}

func (w jsonWriter) Write() error {
	file, err := openOrCreateFile(w.outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(w.items)
}
