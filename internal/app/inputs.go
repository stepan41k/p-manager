package app

import (
	"charm.land/bubbles/v2/textinput"
	"github.com/stepan41k/p-manager/internal/crypto"
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
		{Category: "SETUP", Name: "Next", Binding: &m.keys.Setup.Next},
		{Category: "SETUP", Name: "Previous", Binding: &m.keys.Setup.Previous},
		{Category: "VAULT", Name: "Create", Binding: &m.keys.Vault.Create},
		{Category: "VAULT", Name: "Delete", Binding: &m.keys.Vault.Delete},
		{Category: "VAULT", Name: "Edit", Binding: &m.keys.Vault.Edit},
		{Category: "DETAILS", Name: "Copy", Binding: &m.keys.Details.Copy},
		{Category: "DETAILS", Name: "View", Binding: &m.keys.Details.View},
		{Category: "CREATE", Name: "Next", Binding: &m.keys.Create.Next},
		{Category: "CREATE", Name: "Previous", Binding: &m.keys.Create.Previous},
		{Category: "CREATE", Name: "Generate", Binding: &m.keys.Create.Generate},
		{Category: "EDIT", Name: "Next", Binding: &m.keys.Edit.Next},
		{Category: "EDIT", Name: "Previous", Binding: &m.keys.Edit.Previous},
		{Category: "EDIT", Name: "Generate", Binding: &m.keys.Edit.Generate},
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
	m.inputs = make([]textinput.Model, 4)
	placeholders := []string{"Service", "Email", "Username", "Password"}

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
