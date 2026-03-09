package ui

import (
	"fmt"
)

func (m Model) View() string {
	s := "=== Мой крутой счетчик ===\n\n"
	s += fmt.Sprintf("   Текущее значение: %d\n\n", m.Counter)
	s += "---------------------------\n"
	s += " [Пробел] - увеличить\n"
	s += " [q]      - выйти\n"

	return s
}