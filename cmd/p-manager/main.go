package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/app"
	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/lib/exit"
	"github.com/stepan41k/p-manager/internal/lib/logger/sl"
	"github.com/stepan41k/p-manager/internal/storage/s3"
	"github.com/stepan41k/p-manager/internal/sys"
)

var (
	version   = "0.3.1"
	debugFile = "debug.log"
)

func main() {
	sys.DisableMemoryDumps()

	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("p-manager version %s\n", version)
		os.Exit(0)
	}

	file, err := os.OpenFile(debugFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening degug file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	log := setupLogger(file)

	log.Info("attempting to parse config")

	var initialModel *app.Model
	
	cfg, err := config.MustLoad()
	if err != nil {
		if errors.Is(err, config.ErrNotExists) {
			initialModel = app.NewModel(nil, cfg, nil, log)
			initialModel.SetupInitialInputs()
		} else {
			log.Error("failed to load config:", sl.Err(err))
			fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
			os.Exit(1)
		}
	} else {
		log.Info("config parsed")

		log.Warn("attempting to initialize s3 storage")
		s3Storage, err := s3.New(context.Background(), &cfg.S3Config, log)
		if err != nil {
			log.Error("failed to initialize s3 storage:", sl.Err(err))
			fmt.Fprintf(os.Stderr, "error initializing s3 storage: %v\n", err)
			os.Exit(1)
		}
		log.Info("s3 storage initialized")

		log.Warn("download meta data")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		meta, err := s3Storage.DownloadMeta(ctx)
		if err != nil {
			log.Error("failed to download metadata from S3:", sl.Err(err))
			fmt.Fprintf(os.Stderr, "\n❌ Network Error: Unable to connect to S3 storage.\n")
			fmt.Fprintf(os.Stderr, "   Please check your internet connection and try again.\n\n")
			os.Exit(1)
		}

		log.Info("meta data downloaded")

		initialModel = app.NewModel(s3Storage, cfg, meta, log)
	}

	p := tea.NewProgram(initialModel)

	if _, err = p.Run(); err != nil {
		log.Error("failed to run program", sl.Err(err))
		fmt.Fprintf(os.Stderr, "error running application: %v\n", err)
		os.Exit(1)
	}

	if err = exit.GracefullyStop(log, file); err != nil {
		log.Error("failed to gracefully stop", sl.Err(err))
		fmt.Fprintf(os.Stderr, "error gracefully application stop: %v\n", err)
		os.Exit(1)
	}
}

func setupLogger(w io.Writer) *slog.Logger {
	log := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return log
}
