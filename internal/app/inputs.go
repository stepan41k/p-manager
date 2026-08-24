package app

import (
	"charm.land/bubbles/v2/textinput"
	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/crypto"
	"github.com/zalando/go-keyring"
)

func (m *Model) SetupInitialInputs() {
	m.focusIndex = 0
	m.inputs = make([]textinput.Model, 11)

	labels := []string{
		"S3 Region", "S3 Endpoint", "S3 Bucket", "AWS Access Key", "AWS Secret Key",
		"SMTP Host", "SMTP Port", "Sender Email", "Sender Password",
		"Target Email", "Set Master Password",
	}

	for i := range m.inputs {
		t := textinput.New()
		t.Prompt = ""
		t.SetWidth(40)
		t.Placeholder = labels[i]
		if i == 4 || i == 8 || i == 10 {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '*'
		}
		m.inputs[i] = t
	}
	m.inputs[0].Focus()
}

func (m *Model) setupAuthInput() {
	m.focusIndex = 0
	m.inputs = make([]textinput.Model, 1)

	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.Placeholder = ""
	ti.EchoCharacter = '*'
	ti.Prompt = ""
	ti.SetWidth(40)
	ti.Focus()

	m.inputs[0] = ti
}

func (m *Model) setupOTPInput() {
	m.focusIndex = 0
	m.inputs = make([]textinput.Model, 1)

	t := textinput.New()
	t.Placeholder = "******"
	t.CharLimit = 6
	t.SetWidth(10)
	t.Prompt = ""
	t.Focus()

	m.inputs[0] = t
}

func (m *Model) setupKeymapList() {
	m.keymapIndex = 0
	m.isRebinding = false

	m.bindList = []ConfigurableKey{
		{Category: "COMMON", Name: "Next", Binding: &m.keys.Common.Next},
		{Category: "COMMON", Name: "Previous", Binding: &m.keys.Common.Previous},
		{Category: "COMMON", Name: "Submit", Binding: &m.keys.Common.Submit},
		{Category: "COMMON", Name: "Cancel", Binding: &m.keys.Common.Cancel},
		{Category: "COMMON", Name: "Generate", Binding: &m.keys.Common.Generate},
		{Category: "COMMON", Name: "Quit", Binding: &m.keys.Common.Quit},

		{Category: "VAULT", Name: "Details", Binding: &m.keys.Vault.Details},
		{Category: "VAULT", Name: "Keymaps Config", Binding: &m.keys.Vault.ConfigKeys},
		{Category: "VAULT", Name: "P-Generator Config", Binding: &m.keys.Vault.GenConfig},
		{Category: "VAULT", Name: "Create", Binding: &m.keys.Vault.Create},
		{Category: "VAULT", Name: "Edit", Binding: &m.keys.Vault.Edit},
		{Category: "VAULT", Name: "Delete", Binding: &m.keys.Vault.Delete},
		{Category: "VAULT", Name: "Unauthorize", Binding: &m.keys.Vault.Unauthorize},

		{Category: "P-GENERATOR-CONFIG", Name: "ReduceLength", Binding: &m.keys.GenConfig.ReduceLength},
		{Category: "P-GENERATOR-CONFIG", Name: "IncreaseLength", Binding: &m.keys.GenConfig.IncreaseLength},
		{Category: "P-GENERATOR-CONFIG", Name: "Switch", Binding: &m.keys.GenConfig.Switch},

		{Category: "DETAILS", Name: "Copy", Binding: &m.keys.Details.Copy},
		{Category: "DETAILS", Name: "View", Binding: &m.keys.Details.View},
	}
}

func (m *Model) setupGenConfig() {
	m.genOptIndex = 0

	if m.config != nil && m.config.Generator.Length > 0 {
		m.genOpts = m.config.Generator
	} else {
		m.genOpts = crypto.DefaultGeneratorOptions()
	}

	pass, err := crypto.GeneratePasswordWithOptions(m.genOpts)
	if err == nil {
		m.previewPass = pass
	} else {
		m.previewPass = "error generating preview"
	}
}

func (m *Model) setupFormInputs(values []string) {
	m.focusIndex = 0
	m.inputs = make([]textinput.Model, 5)
	placeholders := []string{"Service", "Email", "Username", "Password", "Notes"}

	for i := range m.inputs {
		t := textinput.New()
		t.SetWidth(40)
		t.Prompt = ""
		t.Placeholder = placeholders[i]

		if len(values) > i {
			if i == 3 {
				t.EchoMode = textinput.EchoPassword
			}
			t.SetValue(values[i])
		}

		m.inputs[i] = t
	}

	m.inputs[0].Focus()
}

func (m *Model) setupSettingsInputs() {
	m.focusIndex = 0
	m.inputs = make([]textinput.Model, 10)

	accKey, _ := keyring.Get("p-manager", "access_key")
	secKey, _ := keyring.Get("p-manager", "secret_key")
	smtpPass, _ := keyring.Get("p-manager", "smtp_password")

	cfg := m.config
	if cfg == nil {
		cfg = &config.Config{}
	}

	values := []string{
		cfg.S3Config.Region,
		cfg.S3Config.Endpoint,
		cfg.S3Config.Bucket,
		accKey,
		secKey,
		cfg.SMTPConfig.SMTPHost,
		cfg.SMTPConfig.SMTPPort,
		cfg.SMTPConfig.SMTPSender,
		smtpPass,
		cfg.SMTPConfig.Email,
	}

	placeholders := []string{
		"S3 Region", "S3 Endpoint", "S3 Bucket", "AWS Access Key", "AWS Secret Key",
		"SMTP Host", "SMTP Port", "Sender Email", "Sender Password", "Target Email",
	}

	for i := range m.inputs {
		t := textinput.New()
		t.SetWidth(40)
		t.Prompt = ""
		t.Placeholder = placeholders[i]
		t.SetValue(values[i])

		if i == 3 || i == 4 || i == 8 {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '*'
		}
		m.inputs[i] = t
	}

	m.inputs[0].Focus()
}
