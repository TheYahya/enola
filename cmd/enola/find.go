package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sherlock-project/enola"
)

type responseMsg enola.Result

func findAndShowResult(username, site string, onlyFound bool) {
	ctx := context.Background()
	sh, err := enola.New(ctx)
	if err != nil {
		panic(err)
	}

	m := model{
		list:             list.New([]list.Item{}, NewDelegate(), 0, 0),
		res:              sh.SetSite(site).Check(username),
		displayOnlyFound: onlyFound,
	}

	m.list.Title = "Socials"
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
