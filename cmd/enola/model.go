package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theyahya/enola"
)

const (
	CheckStyleLightColor      = "#13A10E"
	CheckStyleDarkColor       = "#13A10E"
	TitleStyleLightColor      = "#1a1a1a"
	TitleStyleDarkColor       = "#dddddd"
	DescStyleLightColor       = "#13A10E"
	DescStyleDarkColor        = "#13A10E"
	CloseStyleLightColor      = "#ff0000"
	CloseStyleDarkColor       = "#ff0000"
	TitleNotFoundedLightColor = "#A49FA5"
	TitleNotFoundedDarkColor  = "#777777"
	NotFoundLightColor        = "#ff8f5f"
	NotFoundDarkColor         = "#ffff00"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title string
	desc  string
	found bool
}

func (i item) Title() string {
	status, title, desc := i.renderItem(NewItemStyles().NormalTitle.Copy())
	return fmt.Sprintf("%s %s: %s", status, title, desc)
}

func (i item) renderItem(style lipgloss.Style) (string, string, string) {
	var status, title, desc string
	if i.found {
		status, title, desc = i.renderFoundedItem(style)
	} else {
		status, title, desc = i.renderNotFoundedItem(style)
	}
	return status, title, desc
}

func (i item) renderNotFoundedItem(style lipgloss.Style) (string, string, string) {
	closeStyle := style.Copy().
		Foreground(lipgloss.AdaptiveColor{
			Light: CloseStyleLightColor,
			Dark:  CloseStyleDarkColor,
		})

	titleStyle := style.Copy().
		Foreground(lipgloss.AdaptiveColor{
			Light: TitleNotFoundedLightColor,
			Dark:  TitleNotFoundedDarkColor,
		}).Padding(0, 0, 0, 0)

	notFoundStyle := style.Copy().
		Foreground(lipgloss.AdaptiveColor{
			Light: NotFoundLightColor,
			Dark:  NotFoundDarkColor,
		}).Padding(0, 0, 0, 0)

	status := closeStyle.Render("✗")
	title := titleStyle.Render(i.title)
	desc := notFoundStyle.Render("Not found!")

	return status, title, desc
}

func (i item) renderFoundedItem(style lipgloss.Style) (string, string, string) {
	checkStyle := style.Copy().
		Foreground(lipgloss.AdaptiveColor{
			Light: CheckStyleLightColor,
			Dark:  CheckStyleDarkColor,
		})

	titleStyle := style.Copy().
		Foreground(lipgloss.AdaptiveColor{
			Light: TitleStyleLightColor,
			Dark:  TitleStyleDarkColor,
		}).Bold(true).Padding(0, 0, 0, 0)

	descStyle := NewItemStyles().NormalDesc.Copy().
		Foreground(lipgloss.AdaptiveColor{
			Light: DescStyleLightColor,
			Dark:  DescStyleDarkColor,
		}).Padding(0, 0, 0, 0)

	check := checkStyle.Render("✓")
	title := titleStyle.Render(i.title)
	desc := descStyle.Render(i.desc)

	return check, title, desc
}

func (i item) Description() string { return i.desc }

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
		if msg.Found {
			m.list.InsertItem(0, item{title: msg.Name, desc: msg.URL, found: msg.Found})
			return m, waitForActivity(m.res)
		}
		m.list.InsertItem(m.resCount, item{title: msg.Name, desc: msg.URL, found: msg.Found})
		return m, waitForActivity(m.res)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) updateList(msg responseMsg) (tea.Model, tea.Cmd) {
	m.resCount++
	m.list.InsertItem(m.resCount, item{
		title: msg.Name,
		desc:  msg.URL,
		found: msg.Found,
	})
	return m, waitForActivity(m.res)
}

func (m *model) View() string {
	return docStyle.Render(m.list.View())
}

func waitForActivity(sub <-chan enola.Result) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}
