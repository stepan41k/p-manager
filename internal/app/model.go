package app

import (
	"context"
	"io"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"

	"github.com/atotto/clipboard"
	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/crypto"
	"github.com/stepan41k/p-manager/internal/storage/s3"
)

type sessionState int

const (
	setupState   sessionState = iota // Initial setup
	authState                        // Entering the master password
	otpState                         // Enter OTP
	vaultState                       // Searching for and selecting a entry
	detailsState                     // View entry details
	createState                      // Creating a new entry
	editState                        // Editing the entry
	deleteState                      // Deleting the entry
)

type VaultStorage interface {
	Upload(ctx context.Context, key string, body io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type Model struct {
	// Window size and app state
	width  int
	height int
	state  sessionState

	// Infrastructure
	storage VaultStorage
	config  *config.Config
	log     *slog.Logger

	// Security and Session
	authKey         []byte
	vaultKey        []byte
	meta            s3.Metadata
	expectedOTPHash [32]byte
	otpExpiresAt    time.Time
	otpAttempts     int
	lastActivity    time.Time
	showPassword    bool
	hidePasswordAt  time.Time

	// State of forms and lists
	inputs       []textinput.Model
	focusIndex   int
	vaultList    list.Model
	selectedItem VaultItem
	errorMessage string

	// Design and Hotkeys
	styles Styles
	keys   KeyMap
	help   help.Model
}

func NewModel(s3 *s3.Storage, cfg *config.Config, meta *s3.Metadata, log *slog.Logger) *Model {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.Placeholder = ""
	ti.EchoCharacter = '*'
	ti.SetWidth(40)
	ti.Focus()

	currentStyles := NewStyles(true)

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
		config:    cfg,
		styles:    currentStyles,
		vaultList: vList,
		log:       log,
		keys:      keys,
		help:      help,
	}

	if meta != nil {
		m.meta = *meta
	}

	if startState == setupState {
		m.SetupInitialInputs()
	} else {
		m.setupAuthInput()
	}

	return m
}

func (m *Model) WipeSecrets() {
	m.log.Info("wiped secrets")

	crypto.WipeBytes(m.authKey)
	crypto.WipeBytes(m.vaultKey)

	if m.selectedItem.Password != "" {
		currentContent, err := clipboard.ReadAll()
		if err == nil && currentContent == m.selectedItem.Password {
			_ = clipboard.WriteAll("")
		}
	}

	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}

	m.errorMessage = ""
}
