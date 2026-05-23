package main

import (
	"context"
	"fmt"
	"os"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/theyahya/enola"
	"github.com/theyahya/enola/cmd/exporter"
)

type responseMsg enola.Result

func findAndShowResult(options cmdOptions) {
	ctx := context.Background()
	sh, err := enola.New(ctx)
	if err != nil {
		panic(err)
	}

	resChan, err := sh.SetSite(options.site).Check(options.username)
	if err != nil {
		fmt.Println("Error running program: ", err)
		os.Exit(1)
	}

	m := model{
		list: list.New(
			[]list.Item{},
			NewDelegate(false),
			0,
			0,
		),
		res: resChan,
	}

	m.list.Title = "Socials"
	p := tea.NewProgram(&m)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program: ", err)
		os.Exit(1)
	}

	defer func() {
		if options.outputPath != "" {
			exportType := exporter.CheckExportType(options.outputPath)

			var writer exporter.Writer = nil
			if exporter.JSON == exportType {
				writer = exporter.JsonWriter{
					OutputPath: options.outputPath,
					Items:      prepareItemsForExport(m.list.Items()),
				}
			} else if exporter.CSV == exportType {
				writer = exporter.CsvWriter{
					OutputPath: options.outputPath,
					Items:      prepareItemsForExport(m.list.Items()),
				}
			}

			if writer != nil {
				writer.Write()
			}
		}
	}()
}

func prepareItemsForExport(items []list.Item) []exporter.Item {
	var ret []exporter.Item

	for _, value := range items {
		if itemValue, ok := value.(item); ok {
			ret = append(ret, exporter.Item{Title: itemValue.title, URL: itemValue.desc, Found: itemValue.found})
		}
	}

	return ret
}
