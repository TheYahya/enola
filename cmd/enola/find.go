package main

import (
	"context"
	"fmt"
	"os"

	"github.com/theyahya/enola"
	"github.com/theyahya/enola/cmd/enola/internal/tui"
	"github.com/theyahya/enola/internal/export"
)

func findAndShowResult(options cmdOptions) {
	ctx := context.Background()
	e, err := enola.New()
	if err != nil {
		fmt.Println("Error initializing enola:", err)
		os.Exit(1)
	}

	resChan, err := e.SetSite(options.site).Check(ctx, options.username)
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	var items []tui.Item
	if isInteractive() {
		items, err = tui.Run(resChan)
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	} else {
		items = printResults(resChan)
	}

	exportResults(options.outputPath, items)
}

func toExportItems(items []tui.Item) []export.Item {
	out := make([]export.Item, len(items))
	for i, item := range items {
		out[i] = export.Item{
			Title: item.Title,
			URL:   item.URL,
			Found: item.Found,
		}
	}
	return out
}
