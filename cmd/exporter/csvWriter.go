package exporter

import (
	"encoding/csv"
	"fmt"
	"strconv"
)

type CsvWriter struct {
	OutputPath string
	Items      []Item
}

func (writer CsvWriter) Write() {
	file, err := OpenOrCreateFile(writer.OutputPath)
	if err != nil {
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	err = csvWriter.Write([]string{"title", "url", "found"})
	if err != nil {
		fmt.Println("Error while saving content to csv file")
		return
	}

	for _, item := range writer.Items {
		row := []string{item.Title, item.URL, strconv.FormatBool(item.Found)}
		err = csvWriter.Write(row)
		if err != nil {
			fmt.Println("Error while saving content to csv file")
			return
		}
	}
}
