package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sherlock-project/enola"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
	found       bool
}

func (i item) Title() string {
	style := NewItemStyles().NormalTitle.Copy()
	if i.found {
		checkStyle := style.Copy().
			Foreground(lipgloss.AdaptiveColor{Light: "#13A10E", Dark: "#13A10E"})

		titleStyle := style.Copy().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
			Bold(true).
			Padding(0, 0, 0, 0)

		descStyle := NewItemStyles().NormalDesc.Copy().
			Foreground(lipgloss.AdaptiveColor{Light: "#13A10E", Dark: "#13A10E"}).
			Padding(0, 0, 0, 0)

		check := checkStyle.Render("✓")
		title := titleStyle.Render(i.title)
		desc := descStyle.Render(i.desc)

		return fmt.Sprintf("%s %s: %s", check, title, desc)
	}

	closeStyle := style.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "#ff0000", Dark: "#ff0000"})
	titleStyle := style.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 0, 0)

	notFoundStyle := style.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "#ff8f5f", Dark: "#ffff00"}).
		Padding(0, 0, 0, 0)

	close := closeStyle.Render("✗")
	title := titleStyle.Render(i.title)
	notFound := notFoundStyle.Render("Not found!")

	return fmt.Sprintf("%s %s: %s", close, title, notFound)
}

func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

type model struct {
	list     list.Model
	res      <-chan enola.Result
	resCount int
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		waitForActivity(m.res),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case responseMsg:
		m.resCount++
		m.list.InsertItem(m.resCount, item{title: msg.Name, desc: msg.URL, found: msg.Found})
		return m, waitForActivity(m.res)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	return docStyle.Render(m.list.View())
}

func waitForActivity(sub <-chan enola.Result) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}
