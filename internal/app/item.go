package app

import (
	"fmt"

	"charm.land/bubbles/v2/list"
)

type VaultItem struct {
	Resource    string `json:"resource"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Note        string `json:"note,omitempty"`
	IsDuplicate bool   `json:"-"`
}

func GetVaultItem(it list.Item) (VaultItem, bool) {
	if v, ok := it.(VaultItem); ok {
		return v, true
	}
	if v, ok := it.(*VaultItem); ok && v != nil {
		return *v, true
	}
	return VaultItem{}, false
}

func (i VaultItem) FilterValue() string { return fmt.Sprintf("%s %s", i.Resource, i.Note) }
func (i VaultItem) Title() string {
	if i.IsDuplicate {
		return i.Resource + " ⚠️"
	}
	return i.Resource
}
func (i VaultItem) Description() string { return i.Email + " " + i.Username }
