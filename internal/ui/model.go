package ui

import (
	"context"
	"io"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"

	"github.com/stepan41k/p-manager/internal/storage/s3"
	"github.com/stepan41k/p-manager/internal/ui/styles"
)

type sessionState int

const (
	authState    sessionState = iota // Ввод мастер-пароля
	vaultState                       // Поиск и выбор аккаунта
	stateDetails                     // Просмотр деталей или добавление нового пароля
	createState                      // Создание нового пароля
)

type VaultStorage interface {
	Upload(ctx context.Context, key string, body io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type Model struct {
	storage      VaultStorage
	state        sessionState
	passInput    textinput.Model
	vaultList    list.Model
	selectedItem VaultItem
	choice       string
	styles       styles.Styles
	errorMessage string
	quitting     bool
	inputs       []textinput.Model
	focusIndex   int
	// errorMsg string
}

func NewModel(s3 *s3.Storage) *Model {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.Placeholder = "Введите мастер-пароль"
	ti.EchoCharacter = '*'
	ti.SetWidth(30)
	ti.Focus()

	defaultDelegate := list.NewDefaultDelegate()

	vList := list.New([]list.Item{}, defaultDelegate, 0, 0)
	vList.Title = "Загрузка..."

	m := &Model{
		state:     authState,
		storage:   s3,
		passInput: ti,
		vaultList: vList,
	}

	m.UpdateStyles(true)

	return m
}

func (m *Model) setupInputs() {
	m.focusIndex = 0
	
	m.inputs = make([]textinput.Model, 4)
	
	
	for i := range m.inputs {
			t := textinput.New()
			t.Width = 30
			
			switch i {
			case 0:
				t.Placeholder = "Название сервиса"
				t.Focus()
			case 1:
				t.Placeholder = "Логин / Email"
			case 2:
				t.Placeholder = "Пароль"
				// Для поля пароля можно включить скрытие, 
				// но при создании часто удобнее видеть, что вводишь
			case 3:
				
			}
			m.inputs[i] = t
		}
		
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Название сервиса"
	m.inputs[0].Focus()

	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Имя пользователя"

	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "Email"

	m.inputs[3] = textinput.New()
	m.inputs[3].Placeholder = "Пароль (нажмите g для генерации)"
}

func (m Model) UpdateStyles(isDark bool) {
	m.styles = styles.NewStyles(isDark)
	m.vaultList.Styles.Title = m.styles.Title
	m.vaultList.Styles.PaginationStyle = m.styles.Pagination
	m.vaultList.Styles.HelpStyle = m.styles.Help
	m.vaultList.SetDelegate(ItemDelegate{styles: m.styles})
}
