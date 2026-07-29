package ui

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Auth    authKeyMap
	Vault   vaultKeyMap
	Details detailsKeyMap
	Form    formKeyMap
}

type authKeyMap struct {
	Enter key.Binding
	Quit  key.Binding
}

type vaultKeyMap struct {
	Create  key.Binding
	Details key.Binding
	Edit key.Binding
	Quit    key.Binding
}

type detailsKeyMap struct {
	Back key.Binding
	Copy key.Binding
}

type formKeyMap struct {
	Submit   key.Binding
	Cancel   key.Binding
	Next     key.Binding
	Previous key.Binding
	Generate key.Binding
}

func NewKeyMap() KeyMap {
	return KeyMap{
		Auth: authKeyMap{
			Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "sign in")),
			Quit:  key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl + c", "exit")),
		},
		Vault: vaultKeyMap{
			Create:  key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "create new vault")),
			Details: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "show details")),
			Edit: key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "edit")),
			Quit:    key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "exit")),
		},
		Details: detailsKeyMap{
			Back: key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
			Copy: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")),
		},
		Form: formKeyMap{
			Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "save")),
			Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
			Next:   key.NewBinding(key.WithKeys("tab", "down"), key.WithHelp("tab", "down")),
			Previous:   key.NewBinding(key.WithKeys("shift+tab", "up"), key.WithHelp("shift+tab", "up")),
			Generate:    key.NewBinding(key.WithKeys("ctrl+g"), key.WithHelp("ctrl+g", "generate")),
		},
	}
}

func (k authKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Enter, k.Quit} }
func (k detailsKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Back, k.Copy} }
func (k formKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Submit, k.Generate, k.Cancel} }
