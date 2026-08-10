package ui

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Setup setupKeyMap
	Auth    authKeyMap
	OTP     otpKeyMap
	Vault   vaultKeyMap
	Details detailsKeyMap
	Create  createKeyMap
	Edit    editKeyMap
	Delete  deleteKeyMap
}

type setupKeyMap struct {
	Enter key.Binding
	Quit  key.Binding
}

type authKeyMap struct {
	Enter key.Binding
	Quit  key.Binding
}

type otpKeyMap struct {
	Enter key.Binding
	Quit  key.Binding
}

type vaultKeyMap struct {
	Create  key.Binding
	Details key.Binding
	Edit    key.Binding
	Delete  key.Binding
	Quit    key.Binding
}

type detailsKeyMap struct {
	Back key.Binding
	Copy key.Binding
}

type createKeyMap struct {
	Submit   key.Binding
	Cancel   key.Binding
	Next     key.Binding
	Previous key.Binding
	Generate key.Binding
}

type editKeyMap struct {
	Submit   key.Binding
	Cancel   key.Binding
	Next     key.Binding
	Previous key.Binding
	Generate key.Binding
}

type deleteKeyMap struct {
	Yes key.Binding
	No  key.Binding
}

func NewKeyMap() KeyMap {
	return KeyMap{
		Setup: setupKeyMap{
			Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "sign in")),
			Quit:  key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl + c", "exit")),
		},
		Auth: authKeyMap{
			Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "sign in")),
			Quit:  key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl + c", "exit")),
		},
		OTP: otpKeyMap{
			Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "sign in")),
			Quit:  key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl + c", "exit")),
		},
		Vault: vaultKeyMap{
			Create:  key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "create new vault")),
			Details: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "show details")),
			Edit:    key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "edit")),
			Delete:  key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete")),
			Quit:    key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "exit")),
		},
		Details: detailsKeyMap{
			Back: key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
			Copy: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")),
		},
		Create: createKeyMap{
			Submit:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "save")),
			Cancel:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
			Next:     key.NewBinding(key.WithKeys("tab", "down"), key.WithHelp("tab", "down")),
			Previous: key.NewBinding(key.WithKeys("shift+tab", "up"), key.WithHelp("shift+tab", "up")),
			Generate: key.NewBinding(key.WithKeys("ctrl+g"), key.WithHelp("ctrl+g", "generate")),
		},
		Edit: editKeyMap{
			Submit:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "save")),
			Cancel:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
			Next:     key.NewBinding(key.WithKeys("tab", "down"), key.WithHelp("tab", "down")),
			Previous: key.NewBinding(key.WithKeys("shift+tab", "up"), key.WithHelp("shift+tab", "up")),
			Generate: key.NewBinding(key.WithKeys("ctrl+g"), key.WithHelp("ctrl+g", "generate")),
		},
		Delete: deleteKeyMap{
			Yes: key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
			No:  key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
		},
	}
}

func (k setupKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Enter, k.Quit} }
func (k authKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Enter, k.Quit} }
func (k otpKeyMap) ShortHelp() []key.Binding  { return []key.Binding{k.Enter, k.Quit} }
func (k vaultKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Create, k.Details, k.Edit, k.Quit} }
func (k detailsKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Back, k.Copy} }
func (k createKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Submit, k.Cancel, k.Next, k.Previous, k.Generate} }
func (k editKeyMap) ShortHelp() []key.Binding {	return []key.Binding{k.Submit, k.Cancel, k.Next, k.Previous, k.Generate} }
func (k deleteKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Yes, k.No} }

func (k setupKeyMap) FullHelp() [][]key.Binding    { return [][]key.Binding{k.ShortHelp()} }
func (k authKeyMap) FullHelp() [][]key.Binding    { return [][]key.Binding{k.ShortHelp()} }
func (k otpKeyMap) FullHelp() [][]key.Binding     { return [][]key.Binding{k.ShortHelp()} }
func (k vaultKeyMap) FullHelp() [][]key.Binding   { return [][]key.Binding{k.ShortHelp()} }
func (k detailsKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }
func (k createKeyMap) FullHelp() [][]key.Binding  { return [][]key.Binding{k.ShortHelp()} }
func (k editKeyMap) FullHelp() [][]key.Binding    { return [][]key.Binding{k.ShortHelp()} }
func (k deleteKeyMap) FullHelp() [][]key.Binding  { return [][]key.Binding{k.ShortHelp()} }
