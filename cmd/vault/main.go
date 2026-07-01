package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/lib/logger/sl"
	"github.com/stepan41k/p-manager/internal/storage/s3"
	"github.com/stepan41k/p-manager/internal/ui"
)

func main() {
	file, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0O666)
	if err != nil {
		os.Exit(1)
	}
	defer file.Close()
	log := setupLogger(file)

	log.Info("attempting to parse config")
	cfg, err := config.MustLoad()
	if err != nil {
		log.Error("failed to load config: %w", sl.Err(err))
		os.Exit(1)
	}
	log.Info("config parsed")

	log.Info("attempting to initialize s3 storage")
	s3Storage, err := s3.New(context.Background(), &cfg.S3Config, log)
	if err != nil {
		log.Error("failed to initialize s3 storage: %w", sl.Err(err))
		os.Exit(1)
	}
	log.Info("s3 storage initialized")

	newModel := ui.NewModel(s3Storage, log)

	p := tea.NewProgram(newModel)

	if _, err := p.Run(); err != nil {
		log.Error("failed to run program: %w", sl.Err(err))
		os.Exit(1)
	}

	if err = file.Truncate(0); err != nil {
		log.Error("failed to truncate log file: %w", sl.Err(err))
	}
	
	c := exec.Command("clear")
    c.Stdout = os.Stdout
    _ = c.Run()
}

func setupLogger(w io.Writer) *slog.Logger {
	var log *slog.Logger

	log = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return log
}
