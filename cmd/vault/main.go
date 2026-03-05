package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/wish"
    bm "github.com/charmbracelet/wish/bubbletea"
    lm "github.com/charmbracelet/wish/logging"
    "github.com/charmbracelet/ssh"
)

type Storage interface {
	Save(data []byte) error
	Load() ([]byte, error)
}

type model struct {
    choices  []string
    cursor   int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" { return m, tea.Quit }
    }
    return m, nil
}
func (m model) View() string {
    return "Мой менеджер паролей через SSH!\nНажми 'q' для выхода.\n"
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    m := model{
        choices: []string{"Google", "Github", "Bank"},
    }
    return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func main() {
    s, err := wish.NewServer(
        wish.WithAddress("0.0.0.0:2222"),
        wish.WithHostKeyPath("/Documents/workspace/p-manager/.ssh/term_info_ed25519"),
        wish.WithMiddleware(
            bm.Middleware(teaHandler),
            lm.Middleware(), 
        ),
    )
    if err != nil {
        fmt.Printf("Ошибка сервера: %v\n", err)
        return
    }

    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        if err = s.ListenAndServe(); err != nil {
            fmt.Printf("Ошибка ListenAndServe: %v\n", err)
        }
    }()

    <-done
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    s.Shutdown(ctx)
}