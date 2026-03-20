package ui

import (
	"charm.land/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stepan41k/p-manager/internal/ui/styles"
)

const listHeight = 14

type sessionState int

const (
	authState sessionState = iota
	listState
	addState
)

type Model struct {
	list     list.Model
	choice   string
	styles   styles.Styles
	quitting bool 
	// state sessionState
	// list list.Model
	// masterInput textinput.Model
	// errorMsg string
	// width int
	// height int
}

func (m Model) NewModel() tea.Cmd{
	items := []list.Item{
		Item("Ramen"),
		Item("Tomato Soup"),
		Item("Hamburgers"),
		Item("Cheeseburgers"),
		Item("Currywurst"),
		Item("Okonomiyaki"),
		Item("Pasta"),
		Item("Fillet Mignon"),
		Item("Caviar"),
		Item("Just Wine"),
	}

	const defaultWidth = 20

	l := list.New(items, ItemDelegate{}, defaultWidth, listHeight)
	l.Title = "What do you want for dinner?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	m = Model{list: l}
	m.UpdateStyles(true) // default to dark styles.
	return m
}

func (m *Model) UpdateStyles(isDark bool) {
	m.styles = styles.NewStyles(isDark)
	m.list.Styles.Title = m.styles.title
	m.list.Styles.PaginationStyle = m.styles.pagination
	m.list.Styles.HelpStyle = m.styles.help
	m.list.SetDelegate(itemDelegate{styles: &m.styles})
}