package ui

import (
	"charm.land/bubbles/v2/key"
)

type listKeyMap struct {
	create key.Binding
	copy key.Binding
	delete key.Binding
}

func newListKeyMap() listKeyMap {
	return listKeyMap{
		create: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "create"),
		),
		copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy"),
		),
		delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
	}
}