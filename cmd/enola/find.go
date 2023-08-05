package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theyahya/enola"
)

type responseMsg enola.Result

func findAndShowResult(username, site string) {
	ctx := context.Background()
	sh, err := enola.New(ctx)
	if err != nil {
		panic(err)
	}

	resChan, err := sh.SetSite(site).Check(username)
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

	m.list.Title = "Socials"
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program: ", err)
		os.Exit(1)
	}
}
