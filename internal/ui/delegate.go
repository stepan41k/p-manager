package ui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/ui/styles"
)

type itemDelegate struct {
	styles styles.Styles
}

func (d *itemDelegate) Height() int                             { return 1 }
func (d *itemDelegate) Spacing() int                            { return 0 }
func (d *itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d *itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(VaultItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Resource)

	fn := d.styles.Item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.SelectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}