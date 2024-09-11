package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
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

func NewDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles = NewItemStyles()

	return delegate
}

func NewItemStyles() (style list.DefaultItemStyles) {
	style.NormalTitle = lipgloss.NewStyle().Padding(0, 0, 0, 1)

	style.NormalDesc = style.NormalTitle.
		Foreground(lipgloss.AdaptiveColor{
			Light: NormalDescStyleLightColor,
			Dark:  NormalDescStyleDarkColor,
		}).
		Underline(true).
		Padding(0, 0, 0, 2)

	style.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{
			Light: SelectedBorderStyleLightColor,
			Dark:  SelectedBorderStyleDarkColor,
		}).Padding(0, 0, 0, 0)

	style.SelectedDesc = style.SelectedTitle.
		Foreground(lipgloss.AdaptiveColor{
			Light: SelectedDescStyleLightColor,
			Dark:  SelectedDescStyleDarkColor,
		}).Underline(true).Padding(0, 0, 0, 1)

	style.DimmedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: DimmedTitleStyleLightColor,
			Dark:  DimmedTitleStyleDarkColor,
		}).Padding(0, 0, 0, 1)

	style.DimmedDesc = style.DimmedTitle.
		Foreground(lipgloss.AdaptiveColor{
			Light: DimmedDescStyleLightColor,
			Dark:  DimmedDescStyleDarkColor,
		})

	return style
}
