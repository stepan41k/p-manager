package ui

import (
	"context"
	"io"
	"log/slog"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"

	"github.com/stepan41k/p-manager/internal/storage/s3"
	"github.com/stepan41k/p-manager/internal/ui/styles"
)

type sessionState int

const (
	setupState   sessionState = iota // Первичная настройка
	authState                        // Ввод мастер-пароля
	otpState                         // Ввод OTP
	vaultState                       // Поиск и выбор аккаунта
	detailsState                     // Просмотр деталей или добавление нового пароля
	createState                      // Создание нового пароля
	editState                        // Редактирование записи
	deleteState                      // Удаление записи
)

type VaultStorage interface {
	Upload(ctx context.Context, key string, body io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type Model struct {
	width  int
	height int

	storage      VaultStorage
	state        sessionState
	passInput    textinput.Model
	otpInput     textinput.Model
	vaultList    list.Model
	selectedItem VaultItem
	styles       styles.Styles

	masterKey       []byte
	salt            []byte
	verifier        []byte
	expectedOTPHash [32]byte

	keys KeyMap
	help help.Model

	errorMessage string
	inputs       []textinput.Model
	focusIndex   int
	log          *slog.Logger
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

	keys := NewKeyMap()
	help := help.New()

	var startState sessionState
	
	if s3 == nil {
		startState = setupState
	} else {
		startState = authState
	}
	
	m := &Model{
		state:     startState,
		storage:   s3,
		passInput: ti,

		keys: keys,
		help: help,

		styles:    currentStyles,
		vaultList: vList,
		log:       log,
	}

	return m
}

func (m *Model) SetupInitialInputs() {
	m.focusIndex = 0
	m.inputs = make([]textinput.Model, 7)
	
	labels := []string{"S3 Region", "S3 Endpoint", "S3 Bucket", "AWS Access Key", "AWS Secret Key", "Your Email", "Master Password"}
	
	for i := range m.inputs {
		t := textinput.New()
		t.SetWidth(40)
		t.Placeholder = labels[i]
		if i >= 4 {
			t.EchoMode = textinput.EchoPassword
		}
		m.inputs[i] = t
	}
	m.inputs[0].Focus()
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

func (m *Model) setupOTPInput() {
	ti := textinput.New()
	ti.Placeholder = "******"
	ti.CharLimit = 6
	ti.Focus()
	ti.SetWidth(10)
	ti.Prompt = "Код: "
	m.otpInput = ti
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
