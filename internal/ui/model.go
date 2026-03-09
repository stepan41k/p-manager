package ui

import tea "github.com/charmbracelet/bubbletea"

type sessionState int

const (
	authState sessionState = iota
	listState
	addState
)

type Model struct {
	Counter int
	// state sessionState
	// list list.Model
	// masterInput textinput.Model
	// errorMsg string
	// width int
	// height int
}

func (m Model) Init() tea.Cmd{
	return nil
}