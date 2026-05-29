package tui

import (
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
)

const (
	SelectedBorderStyleLightColor = "#F793FF"
	SelectedBorderStyleDarkColor  = "#AD58B4"
	NormalDescStyleLightColor     = "#A49FA5"
	NormalDescStyleDarkColor      = "#777777"
	SelectedDescStyleLightColor   = "#A49FA5"
	SelectedDescStyleDarkColor    = "#777777"
	DimmedTitleStyleLightColor    = "#A49FA5"
	DimmedTitleStyleDarkColor     = "#777777"
	DimmedDescStyleLightColor     = "#C2B8C2"
	DimmedDescStyleDarkColor      = "#4D4D4D"
)

func NewDelegate(hasDarkBg bool) list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles = NewItemStyles(hasDarkBg)
	return delegate
}

func NewItemStyles(hasDarkBg bool) (style list.DefaultItemStyles) {
	ld := lipgloss.LightDark(hasDarkBg)

	style.NormalTitle = lipgloss.NewStyle().Padding(0, 0, 0, 1)

	style.NormalDesc = style.NormalTitle.
		Foreground(ld(lipgloss.Color(NormalDescStyleLightColor), lipgloss.Color(NormalDescStyleDarkColor))).
		Underline(true).
		Padding(0, 0, 0, 2)

	style.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(ld(lipgloss.Color(SelectedBorderStyleLightColor), lipgloss.Color(SelectedBorderStyleDarkColor))).
		Padding(0, 0, 0, 0)

	style.SelectedDesc = style.SelectedTitle.
		Foreground(ld(lipgloss.Color(SelectedDescStyleLightColor), lipgloss.Color(SelectedDescStyleDarkColor))).
		Underline(true).Padding(0, 0, 0, 1)

	style.DimmedTitle = lipgloss.NewStyle().
		Foreground(ld(lipgloss.Color(DimmedTitleStyleLightColor), lipgloss.Color(DimmedTitleStyleDarkColor))).
		Padding(0, 0, 0, 1)

	style.DimmedDesc = style.DimmedTitle.
		Foreground(ld(lipgloss.Color(DimmedDescStyleLightColor), lipgloss.Color(DimmedDescStyleDarkColor)))

	return style
}
