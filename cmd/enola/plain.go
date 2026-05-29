package main

import (
	"fmt"
	"os"

	"github.com/theyahya/enola"
	"github.com/theyahya/enola/cmd/enola/internal/tui"
	"github.com/theyahya/enola/internal/export"
	"golang.org/x/term"
)

func isInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func printResults(results <-chan enola.Result) []tui.Item {
	var items []tui.Item
	for r := range results {
		items = append(items, tui.Item{Title: r.Name, URL: r.URL, Found: r.Found})
		if r.Found {
			fmt.Printf("[+] %s: %s\n", r.Name, r.URL)
		} else {
			fmt.Printf("[-] %s: not found\n", r.Name)
		}
	}
	return items
}

func exportResults(outputPath string, items []tui.Item) {
	if outputPath == "" {
		return
	}
	writer, err := export.NewWriter(outputPath, toExportItems(items))
	if err != nil {
		fmt.Println("Error preparing export:", err)
		os.Exit(1)
	}
	if err := writer.Write(); err != nil {
		fmt.Println("Error writing export:", err)
		os.Exit(1)
	}
}
