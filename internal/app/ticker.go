package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

func checkInactivityTicker(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return checkInactivityMsg{}
	})
}

func hidePasswordTicker(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return hidePasswordMsg{}
	})
}

func clearClipboardTicker(copiedPassword string, delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return clearClipboardMsg{copiedPassword: copiedPassword}
	})
}

func checkPasswordTicker(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return checkPasswordMsg{}
	})
}
