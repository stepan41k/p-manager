package ui

import (
	"context"
	"io"
	"log/slog"

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
	editState
)

type VaultStorage interface {
	Upload(ctx context.Context, key string, body io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type Model struct {
	width int
	height int
	
	storage      VaultStorage
	masterKey    []byte
	state        sessionState
	passInput    textinput.Model
	vaultList    list.Model
	selectedItem VaultItem
	styles       styles.Styles
	errorMessage string
	inputs       []textinput.Model
	focusIndex   int
	log *slog.Logger
}

func NewModel(s3 *s3.Storage, log *slog.Logger) *Model {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.Placeholder = ""
	ti.EchoCharacter = '*'
	ti.SetWidth(40)
	ti.Focus()

	currentStyles := styles.NewStyles(true)
	
	defaultDelegate := list.NewDefaultDelegate()

	vList := list.New([]list.Item{}, defaultDelegate, 0, 0)
	vList.Styles.Title = currentStyles.Title
	vList.Styles.PaginationStyle = currentStyles.Pagination
	vList.Styles.HelpStyle = currentStyles.Help
	vList.Title = "Loading..."

	m := &Model{
		state:     authState,
		storage:   s3,
		passInput: ti,
		styles: currentStyles,
		vaultList: vList,
		log: log,
	}

	return m
}

func (m *Model) setupInputs() {
	m.focusIndex = 0
	m.inputs = make([]textinput.Model, 4)

	placeholders := []string{"Service", "Email", "Username", "Password"}

    for i := range m.inputs {
        t := textinput.New()
        t.SetWidth(40)

        t.Prompt = ""
        
        t.Placeholder = placeholders[i]
        if i == 0 {
            t.Focus()
        }
        m.inputs[i] = t
    }
	
	m.inputs[0].Focus()
}

func (m *Model) setupEditInputs() {
    m.focusIndex = 0
    m.inputs = make([]textinput.Model, 4)

    values := []string{
        m.selectedItem.Resource,
        m.selectedItem.Email,
        m.selectedItem.Username,
        m.selectedItem.Password,
    }
    
    placeholders := []string{"Service", "Email", "Username", "Password"}

    for i := range m.inputs {
        t := textinput.New()
        t.SetWidth(40)

        t.Prompt = ""
        
        t.Placeholder = placeholders[i]
        t.SetValue(values[i])
        if i == 0 {
            t.Focus()
        }
        m.inputs[i] = t
    }
}