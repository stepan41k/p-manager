package main

import (
	"context"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/lib/logger/sl"
	"github.com/stepan41k/p-manager/internal/storage/s3"
	"github.com/stepan41k/p-manager/internal/ui"
)

func main() {
	log := setupLogger()
	
	cfg, err := config.MustLoad()
	if err != nil {
		log.Error("failed to load config: %w", sl.Err(err))
		os.Exit(1)
	}
	
	s3Storage, err := s3.New(context.Background(), cfg.S3Config)
	if err != nil {
		log.Error("failed to initialize s3 storage: %w", sl.Err(err))
		os.Exit(1)
	}
	
	newModel := ui.NewModel(s3Storage)

	p := tea.NewProgram(newModel)

	if _, err := p.Run(); err != nil {
		log.Error("failed to run program: %w", sl.Err(err))
		os.Exit(1)
	}
}

func setupLogger() *slog.Logger {
	var log *slog.Logger
	
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	return log
}