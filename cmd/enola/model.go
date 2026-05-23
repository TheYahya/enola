package main

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	title     string
	desc      string
	found     bool
	hasDarkBg bool
}

func (i item) Title() string {
	status, title, desc := i.renderItem(NewItemStyles(i.hasDarkBg).NormalTitle)
	return fmt.Sprintf("%s %s: %s", status, title, desc)
}

func (i item) renderItem(style lipgloss.Style) (string, string, string) {
	if i.found {
		return i.renderFoundedItem(style)
	}
	return i.renderNotFoundedItem(style)
}

func (i item) renderNotFoundedItem(style lipgloss.Style) (string, string, string) {
	ld := lipgloss.LightDark(i.hasDarkBg)

	closeStyle := style.Foreground(ld(lipgloss.Color(CloseStyleLightColor), lipgloss.Color(CloseStyleDarkColor)))
	titleStyle := style.Foreground(ld(lipgloss.Color(TitleNotFoundedLightColor), lipgloss.Color(TitleNotFoundedDarkColor))).Padding(0, 0, 0, 0)
	notFoundStyle := style.Foreground(ld(lipgloss.Color(NotFoundLightColor), lipgloss.Color(NotFoundDarkColor))).Padding(0, 0, 0, 0)

	return closeStyle.Render("✗"), titleStyle.Render(i.title), notFoundStyle.Render("Not found!")
}

func (i item) renderFoundedItem(style lipgloss.Style) (string, string, string) {
	ld := lipgloss.LightDark(i.hasDarkBg)

	checkStyle := style.Foreground(ld(lipgloss.Color(CheckStyleLightColor), lipgloss.Color(CheckStyleDarkColor)))
	titleStyle := style.Foreground(ld(lipgloss.Color(TitleStyleLightColor), lipgloss.Color(TitleStyleDarkColor))).Bold(true).Padding(0, 0, 0, 0)
	descStyle := NewItemStyles(i.hasDarkBg).NormalDesc.Foreground(ld(lipgloss.Color(DescStyleLightColor), lipgloss.Color(DescStyleDarkColor))).Padding(0, 0, 0, 0)

	return checkStyle.Render("✓"), titleStyle.Render(i.title), descStyle.Render(i.desc)
}

func (i item) Description() string { return i.desc }

func (i item) FilterValue() string { return i.title }

type model struct {
	list      list.Model
	res       <-chan enola.Result
	resCount  int
	hasDarkBg bool
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		waitForActivity(m.res),
		tea.RequestBackgroundColor,
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.BackgroundColorMsg:
		m.hasDarkBg = msg.IsDark()
		m.list.SetDelegate(NewDelegate(m.hasDarkBg))
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case responseMsg:
		m.resCount++
		it := item{title: msg.Name, desc: msg.URL, found: msg.Found, hasDarkBg: m.hasDarkBg}
		if msg.Found {
			m.list.InsertItem(0, it)
		} else {
			m.list.InsertItem(m.resCount, it)
		}
		return m, waitForActivity(m.res)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) updateList(msg responseMsg) (tea.Model, tea.Cmd) {
	m.resCount++
	m.list.InsertItem(m.resCount, item{
		title:     msg.Name,
		desc:      msg.URL,
		found:     msg.Found,
		hasDarkBg: m.hasDarkBg,
	})
	return m, waitForActivity(m.res)
}

func (m *model) View() tea.View {
	v := tea.NewView(docStyle.Render(m.list.View()))
	v.AltScreen = true
	return v
}

func waitForActivity(sub <-chan enola.Result) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}
