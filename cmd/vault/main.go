package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/ui"
)

func main() {

	newModel := ui.NewModel()

	p := tea.NewProgram(newModel)

	if _, err := p.Run(); err != nil {
		fmt.Printf("error")
		os.Exit(1)
	}
}