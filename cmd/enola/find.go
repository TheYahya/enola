package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theyahya/enola"
)

type responseMsg enola.Result

func findAndShowResult(username string, option *elonaCommandOption) {
	ctx := context.Background()
	sh, err := enola.New(ctx)
	if err != nil {
		panic(err)
	}

	resChan, err := sh.SetSite(option.site).Check(username)
	if err != nil {
		fmt.Println("Error running program: ", err)
		os.Exit(1)
	}

	m := model{
		list: list.New(
			[]list.Item{},
			NewDelegate(),
			0,
			0,
		),
		res: resChan,
	}

	if option.outputPath != "" {
		defer func(m *model, outputPath string, filename string, printFound bool) {
			err := writeOutput(m, outputPath, filename, printFound)
			if err != nil {
				fmt.Println("Error : ", enola.ErrWritingOutputFailed)
				os.Exit(1)
			}
		}(&m, option.outputPath, username, option.printFound)
	}

	m.list.Title = "Socials"
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program: ", err)
		os.Exit(1)
	}
}

func writeOutput(m *model, outputPath string, filename string, printFound bool) error {
	outputFile, err := getOutputfile(outputPath, filename)
	defer outputFile.Close()
	if err != nil {
		return err
	}

	for _, value := range m.list.Items() {
		if !printFound || value.(item).found {
			_, err := outputFile.WriteString(fmt.Sprintln(value))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getOutputfile(outputPath string, filename string) (*os.File, error) {
	fullPath, err := filepath.Abs(outputPath)
	if err != nil {
		fmt.Println("Error resolving absolute path:", err)
		return nil, err
	}

	if _, err := os.Stat(filepath.Join(fullPath, filename)); os.IsNotExist(err) {
		file, err := os.Create(filepath.Join(fullPath, filename))

		err = os.MkdirAll(fullPath, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating path:", err)
			return nil, err
		}

		if err != nil {
			fmt.Println("Error creating file:", err)
			os.Exit(1)
		}
		return file, nil
	} else if err != nil {
		fmt.Println("Error checking file existence:", err)
		return nil, err
	} else {
		err := os.Remove(filepath.Join(fullPath, filename))
		if err != nil {
			fmt.Println("Error removing existing file:", err)
			return nil, err
		}

		file, err := os.Create(filepath.Join(fullPath, filename))
		if err != nil {
			fmt.Println("Error creating file:", err)
			return nil, err
		}
		return file, nil
	}
}
