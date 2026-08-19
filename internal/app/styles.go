package app

import (
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
)

type Styles struct {
	Title        lipgloss.Style
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	Pagination   lipgloss.Style
	Help         lipgloss.Style
	QuitText     lipgloss.Style
	Card         lipgloss.Style
	DetailLabel  lipgloss.Style
	DetailValue  lipgloss.Style
	DetailKey    lipgloss.Style
}

func NewStyles(darkBG bool) Styles {
	var s Styles
	s.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Padding(0, 1).
		MarginLeft(2).
		Bold(true)

	s.Item = lipgloss.NewStyle().PaddingLeft(4)
	s.SelectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("120"))
	s.Pagination = list.DefaultStyles(darkBG).PaginationStyle.PaddingLeft(4)
	s.Help = list.DefaultStyles(darkBG).HelpStyle.PaddingLeft(4).PaddingBottom(1)
	s.QuitText = lipgloss.NewStyle().Margin(1, 0, 2, 4)

	s.Card = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1, 2).
		Margin(1, 2).
		Width(80)

	s.DetailLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Bold(true).
		Width(10)

	s.DetailValue = lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	s.DetailKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true)

	return s
}
