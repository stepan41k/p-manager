package app

type VaultItem struct {
	Resource string `json:"resource"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (i VaultItem) FilterValue() string { return i.Resource }
func (i VaultItem) Title() string       { return i.Resource }
func (i VaultItem) Description() string { return i.Email + " " + i.Username }
