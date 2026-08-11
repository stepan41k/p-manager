package main

import (
	"context"
	"errors"
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
	file, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		os.Exit(1)
	}
	defer file.Close()

	log := setupLogger(file)

	log.Info("attempting to parse config")

	cfg, err := config.MustLoad()

	var initialModel *ui.Model
	if err != nil {
		if errors.Is(err, config.ErrNotExists) {
			initialModel = ui.NewModel(nil, cfg, nil, nil, log)
			initialModel.SetupInitialInputs()
		} else {
			log.Error("failed to load config:", sl.Err(err))
			os.Exit(1)
		}
	} else {
		log.Info("attempting to initialize s3 storage")

		s3Storage, err := s3.New(context.Background(), &cfg.S3Config, log)
		if err != nil {
			log.Error("failed to initialize s3 storage:", sl.Err(err))
			os.Exit(1)
		}

		log.Info("s3 storage initialized")

		log.Warn("download meta data")

		meta, err := s3Storage.DownloadMeta(context.TODO())
		if err != nil {
			log.Error("failed to download metadata from S3:", sl.Err(err))
			os.Exit(1)
		}

		log.Info("meta data downloaded")
		
		initialModel = ui.NewModel(s3Storage, cfg, meta.Salt, meta.Verifier, log)
	}

	log.Info("config parsed")

	p := tea.NewProgram(initialModel)

	if _, err = p.Run(); err != nil {
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
	log := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return log
}
