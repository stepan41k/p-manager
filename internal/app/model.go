package app

import (
	"context"
	"io"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"

	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/crypto"
	"github.com/stepan41k/p-manager/internal/storage/s3"
)

type sessionState int

const (
	setupState            sessionState = iota // Initial setup
	authState                                 // Entering the master password
	otpState                                  // Enter OTP
	vaultState                                // Searching for and selecting a entry
	settingsState                             // Application settings
	customizeKeymapsState                     // View and customize keymaps
	genConfigState                            // View and customize password generator
	detailsState                              // View entry details
	createState                               // Creating a new entry
	editState                                 // Editing the entry
	deleteState                               // Deleting the entry
)

type VaultStorage interface {
	Upload(ctx context.Context, key string, body io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type ConfigurableKey struct {
	Category string
	Name     string
	Binding  *key.Binding
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
	authAttempts    int
	maskPasswordAt  time.Time

	// State of forms and lists
	inputs       []textinput.Model
	focusIndex   int
	vaultList    list.Model
	selectedItem VaultItem
	errorMessage string

	// Design and Keymaps
	styles      Styles
	keys        KeyMap
	help        help.Model
	keymapIndex int
	isRebinding bool
	bindList    []ConfigurableKey

	// Generator Options
	genOpts     crypto.GeneratorOptions
	genOptIndex int
	previewPass string
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

	m.setupKeymapList()
	m.applyCustomKeys()

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
	crypto.WipeBytes(m.authKey)
	crypto.WipeBytes(m.vaultKey)

	m.selectedItem = VaultItem{}
	m.expectedOTPHash = [32]byte{}
	m.otpAttempts = 0
	m.authAttempts = 0

	m.setupAuthInput()
	m.errorMessage = ""
}

func (m *Model) WipeAll() {
	m.WipeSecrets()

	if len(m.meta.Salt) > 0 {
		crypto.WipeBytes(m.meta.Salt)
	}
	if len(m.meta.Verifier) > 0 {
		crypto.WipeBytes(m.meta.Verifier)
	}
	m.meta = s3.Metadata{}
}