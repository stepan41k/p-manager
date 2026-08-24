package app

import (
	"fmt"
	"slices"

	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Common       commonKeyMap
	Setup        setupKeyMap
	Auth         authKeyMap
	OTP          otpKeyMap
	Vault        vaultKeyMap
	KeyMapConfig keyMapConfig
	GenConfig    genConfigKeyMap
	Settings     settingsKeyMap
	Details      detailsKeyMap
	Create       createKeyMap
	Edit         editKeyMap
	Delete       deleteKeyMap
}

type commonKeyMap struct {
	Next     key.Binding
	Previous key.Binding
	Generate key.Binding
	Submit   key.Binding
	Cancel   key.Binding
	Quit     key.Binding
}

type setupKeyMap struct {
	*commonKeyMap
}

type authKeyMap struct {
	*commonKeyMap
}

type otpKeyMap struct {
	*commonKeyMap
}

type vaultKeyMap struct {
	*commonKeyMap
	ConfigKeys  key.Binding
	GenConfig   key.Binding
	Settings    key.Binding
	Create      key.Binding
	Details     key.Binding
	Edit        key.Binding
	Delete      key.Binding
	Unauthorize key.Binding
}

type keyMapConfig struct {
	*commonKeyMap
}

type genConfigKeyMap struct {
	*commonKeyMap
	ReduceLength   key.Binding
	IncreaseLength key.Binding
	Switch         key.Binding
}

type settingsKeyMap struct {
	*commonKeyMap
	RevokeDevices key.Binding
}

type detailsKeyMap struct {
	*commonKeyMap
	Copy key.Binding
	View key.Binding
}

type createKeyMap struct {
	*commonKeyMap
}

type editKeyMap struct {
	*commonKeyMap
}

type deleteKeyMap struct {
	Yes key.Binding
	No  key.Binding
}

func NewKeyMap() KeyMap {
	commonKeyMap := commonKeyMap{
		Next:     key.NewBinding(key.WithKeys("down", "tab"), key.WithHelp("↓ | tab", "down")),
		Previous: key.NewBinding(key.WithKeys("up", "shift+tab"), key.WithHelp("↑ | shift+tab", "up")),
		Generate: key.NewBinding(key.WithKeys("ctrl+g"), key.WithHelp("ctrl+g", "generate")),
		Submit:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
		Cancel:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Quit:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl + c", "exit")),
	}

	return KeyMap{
		Common: commonKeyMap,
		Setup: setupKeyMap{
			commonKeyMap: &commonKeyMap,
		},
		Auth: authKeyMap{
			commonKeyMap: &commonKeyMap,
		},
		OTP: otpKeyMap{
			commonKeyMap: &commonKeyMap,
		},
		Vault: vaultKeyMap{
			commonKeyMap: &commonKeyMap,
			ConfigKeys:   key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "customize keymaps")),
			GenConfig:    key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "customize password generator")),
			Settings:     key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "user settings")),
			Create:       key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "create new vault")),
			Details:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "show details")),
			Edit:         key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "edit")),
			Delete:       key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete")),
			Unauthorize:  key.NewBinding(key.WithKeys("ctrl+q"), key.WithHelp("ctrl+q", "unathorize")),
		},
		KeyMapConfig: keyMapConfig{
			commonKeyMap: &commonKeyMap,
		},

		Settings: settingsKeyMap{
			commonKeyMap: &commonKeyMap,
			RevokeDevices: key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "revoke all 2FA devices")),
		},

		GenConfig: genConfigKeyMap{
			commonKeyMap:   &commonKeyMap,
			ReduceLength:   key.NewBinding(key.WithKeys("left"), key.WithHelp("left", "←")),
			IncreaseLength: key.NewBinding(key.WithKeys("right"), key.WithHelp("right", "→")),
			Switch:         key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "switch")),
		},
		Details: detailsKeyMap{
			commonKeyMap: &commonKeyMap,
			Copy:         key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")),
			View:         key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view")),
		},
		Create: createKeyMap{
			commonKeyMap: &commonKeyMap,
		},
		Edit: editKeyMap{
			commonKeyMap: &commonKeyMap,
		},
		Delete: deleteKeyMap{
			Yes: key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
			No:  key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
		},
	}
}

func (k commonKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Previous, k.Submit, k.Generate, k.Cancel, k.Quit}
}
func (k setupKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Next, k.Previous, k.Quit}
}
func (k authKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Submit, k.Quit} }
func (k otpKeyMap) ShortHelp() []key.Binding  { return []key.Binding{k.Submit, k.Quit} }
func (k vaultKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ConfigKeys, k.GenConfig, k.Settings, k.Create, k.Details, k.Edit, k.Delete, k.Unauthorize, k.Quit}
}
func (k keyMapConfig) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Next, k.Previous, k.Quit}
}
func (k genConfigKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Previous, k.ReduceLength, k.IncreaseLength, k.Switch, k.Generate, k.Cancel}
}
func (k detailsKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Cancel, k.Copy, k.View} }
func (k createKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Cancel, k.Next, k.Previous, k.Generate}
}
func (k editKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Cancel, k.Next, k.Previous, k.Generate}
}
func (k settingsKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Next, k.Previous, k.RevokeDevices, k.Submit, k.Cancel} }
func (k deleteKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Yes, k.No} }

func (k commonKeyMap) FullHelp() [][]key.Binding    { return [][]key.Binding{k.ShortHelp()} }
func (k setupKeyMap) FullHelp() [][]key.Binding     { return [][]key.Binding{k.ShortHelp()} }
func (k authKeyMap) FullHelp() [][]key.Binding      { return [][]key.Binding{k.ShortHelp()} }
func (k otpKeyMap) FullHelp() [][]key.Binding       { return [][]key.Binding{k.ShortHelp()} }
func (k vaultKeyMap) FullHelp() [][]key.Binding     { return [][]key.Binding{k.ShortHelp()} }
func (k keyMapConfig) FullHelp() [][]key.Binding    { return [][]key.Binding{k.ShortHelp()} }
func (k genConfigKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }
func (k detailsKeyMap) FullHelp() [][]key.Binding   { return [][]key.Binding{k.ShortHelp()} }
func (k createKeyMap) FullHelp() [][]key.Binding    { return [][]key.Binding{k.ShortHelp()} }
func (k editKeyMap) FullHelp() [][]key.Binding      { return [][]key.Binding{k.ShortHelp()} }
func (k deleteKeyMap) FullHelp() [][]key.Binding    { return [][]key.Binding{k.ShortHelp()} }
func (k settingsKeyMap) FullHelp() [][]key.Binding    { return [][]key.Binding{k.ShortHelp()} }

func (m *Model) areCategoriesConflicting(cat1, name1, cat2, name2 string) bool {
	if cat1 == cat2 && name1 == name2 {
		return false
	}

	if cat1 == cat2 {
		return true
	}

	if name1 == "Quit" || name2 == "Quit" {
		return true
	}

	return false
}

var reservedVaultKeys = map[string]string{
	"down":     "List Navigation (Down)",
	"up":       "List Navigation (Up)",
	"j":        "List Navigation (Down)",
	"k":        "List Navigation (Up)",
	"pageup":   "List Page Up",
	"pagedown": "List Page Down",
	"home":     "List Home",
	"end":      "List End",
	"/":        "List Search/Filter",
}

func (m *Model) findKeyConflict(newKey string, targetIndex int) string {
	targetItem := m.bindList[targetIndex]

	if targetItem.Category == "VAULT" {
		if reservedAction, isReserved := reservedVaultKeys[newKey]; isReserved {
			return fmt.Sprintf("Key '%s' is reserved for '%s'!", newKey, reservedAction)
		}
	}

	for i, item := range m.bindList {
		if i == targetIndex {
			continue
		}

		if m.areCategoriesConflicting(targetItem.Category, targetItem.Name, item.Category, item.Name) {
			if item.Binding != nil {
				if slices.Contains(item.Binding.Keys(), newKey) {
					return fmt.Sprintf("Key '%s' is used for '%s' in %s!", newKey, item.Name, item.Category)
				}
			}
		}
	}

	return ""
}

func (m *Model) exportCustomKeys() map[string]string {
	customKeys := make(map[string]string)

	for _, item := range m.bindList {
		if item.Binding != nil && len(item.Binding.Keys()) > 0 {
			keyID := fmt.Sprintf("%s.%s", item.Category, item.Name)
			customKeys[keyID] = item.Binding.Keys()[0]
		}
	}
	return customKeys
}

func (m *Model) applyCustomKeys() {
	if m.config == nil || len(m.config.Keymaps) == 0 {
		return
	}

	m.setupKeymapList()

	for _, item := range m.bindList {
		keyID := fmt.Sprintf("%s.%s", item.Category, item.Name)
		if savedKey, ok := m.config.Keymaps[keyID]; ok && savedKey != "" {
			item.Binding.SetKeys(savedKey)
			desc := item.Binding.Help().Desc
			item.Binding.SetHelp(savedKey, desc)
		}
	}
}
