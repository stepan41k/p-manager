package app

import "fmt"

type VaultItem struct {
	Resource string `json:"resource"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Note     string `json:"note,omitempty"`
}

func (i VaultItem) FilterValue() string { return fmt.Sprintf("%s %s", i.Resource, i.Note) }
func (i VaultItem) Title() string       { return i.Resource }
func (i VaultItem) Description() string { return i.Email + " " + i.Username }
