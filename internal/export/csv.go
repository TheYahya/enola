package export

import (
	"encoding/csv"
	"strconv"
)

type csvWriter struct {
	outputPath string
	items      []Item
}

func (w csvWriter) Write() error {
	file, err := openOrCreateFile(w.outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"title", "url", "found"}); err != nil {
		return err
	}
	for _, item := range w.items {
		row := []string{item.Title, item.URL, strconv.FormatBool(item.Found)}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return writer.Error()
}
