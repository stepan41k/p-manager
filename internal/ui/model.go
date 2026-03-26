package ui

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	// tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/ui/styles"
)

const (
	listHeight = 14
	defaultWidth = 20
)

type sessionState int

const (
	authState sessionState = iota	// Ввод мастер-пароля
	vaultState						// Поиск и выбор аккаунта
	entryState						// Просмотр деталей или добавление нового пароля
)

type Model struct {
	state sessionState
	passInput textinput.Model
	vaultList     list.Model
	choice   string
	styles   styles.Styles
	errorMessage string
	quitting bool 
	// errorMsg string
}

func NewModel() *Model {
	// items := []list.Item{
	// 	// Item("Google"),
	// 	// Item("Yandex"),
	// 	// Item("Lamoda"),
	// 	// Item("Ozon"),
	// 	// Item("Wildberries"),
	// 	// Item("Avito"),
	// 	// Item("Uber"),
	// 	// Item("GitHub"),
	// 	// Item("Steam"),
	// 	// Item("Music"),
	// }

	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.Placeholder = "Введите мастер-пароль"
	ti.EchoCharacter = '*'
	ti.Focus()
	

	defaultDelegate := list.NewDefaultDelegate()

	vList := list.New([]list.Item{}, defaultDelegate, 0, 0)
	vList.Title = "Загрузка..."

	// l := list.New(items, ItemDelegate{}, defaultWidth, listHeight)
	// l.Title = "Input password:"
	// l.SetShowStatusBar(false)
	// l.SetFilteringEnabled(false)

	// m := &Model{state: authState, list: l}
	// m.UpdateStyles(true)

	m := &Model{
		state: authState,
		passInput: ti,
		vaultList: vList,
	}

	m.UpdateStyles(true)

	return m
}

func (m Model) UpdateStyles(isDark bool) {
	m.styles = styles.NewStyles(isDark)
	m.vaultList.Styles.Title = m.styles.Title
	m.vaultList.Styles.PaginationStyle = m.styles.Pagination
	m.vaultList.Styles.HelpStyle = m.styles.Help
	m.vaultList.SetDelegate(ItemDelegate{styles: m.styles})
}