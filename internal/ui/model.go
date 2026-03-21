package ui

import (
	"charm.land/bubbles/v2/list"
	// tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/ui/styles"
)

const listHeight = 14

type sessionState int

const (
	authState sessionState = iota
	vaultState
	entryState
)

type Model struct {
	State sessionState
	List     list.Model
	Choice   string
	Styles   styles.Styles
	Quitting bool 
	// state sessionState
	// list list.Model
	// masterInput textinput.Model
	// errorMsg string
	// width int
	// height int
}

func NewModel() *Model {
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

	m := &Model{List: l}
	m.UpdateStyles(true)
	return m
}

func (m *Model) UpdateStyles(isDark bool) {
	m.Styles = styles.NewStyles(isDark)
	m.List.Styles.Title = m.Styles.Title
	m.List.Styles.PaginationStyle = m.Styles.Pagination
	m.List.Styles.HelpStyle = m.Styles.Help
	m.List.SetDelegate(ItemDelegate{styles: m.Styles})
}