package app

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type itemDelegate struct {
	styles Styles
}

func (d *itemDelegate) Height() int                             { return 1 }
func (d *itemDelegate) Spacing() int                            { return 0 }
func (d *itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := GetVaultItem(listItem)
	if !ok {
		return
	}

	title := i.Resource

	if i.IsDuplicate {
		warningBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true).
			Render(" ⚠️")

		title = title + warningBadge
	}

	str := fmt.Sprintf("%s\n%s %s", title, i.Email, i.Username)

	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedItem.Render("> "+str))
	} else {
		fmt.Fprint(w, d.styles.Item.Render("  "+str))
	}
}
