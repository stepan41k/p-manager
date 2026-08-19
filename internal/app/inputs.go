package app

import "charm.land/bubbles/v2/textinput"

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
			t.SetValue(values[i])
		}

		m.inputs[i] = t
	}

	m.inputs[0].Focus()
}
