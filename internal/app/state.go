package app

func (m *Model) SetState(s sessionState) {
	m.state = s
	m.errorMessage = ""
}
