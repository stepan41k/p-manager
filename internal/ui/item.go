package ui

type VaultItem struct {
	resource string
	email    string
	username string
	password string
}

func (i VaultItem) FilterValue() string { return i.resource }
func (i VaultItem) Title() string       { return i.resource }
func (i VaultItem) Description() string { return i.email + i.username }
