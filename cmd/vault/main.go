package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/app"
	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/lib/logger/sl"
	"github.com/stepan41k/p-manager/internal/storage/s3"
	"github.com/stepan41k/p-manager/internal/sys"
)

var (
	version = "dev" 
	debugFile = "debug.log"
)

func main() {
	sys.DisableMemoryDumps()
	
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
        fmt.Printf("p-manager version %s\n", version)
        os.Exit(0)
    }
    
	var errorMessage string
	
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic:", err)
		}
	}()
	
	file, err := os.OpenFile(debugFile, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0o666)
	if err != nil {
		errorMessage = fmt.Sprintf("failed to open debug file: %s", err.Error())
		panic(errorMessage)
	}
	defer file.Close()

	log := setupLogger(file)

	log.Info("attempting to parse config")
	cfg, err := config.MustLoad()

	var initialModel *app.Model
	if err != nil {
		if errors.Is(err, config.ErrNotExists) {
			initialModel = app.NewModel(nil, cfg, nil, log)
			initialModel.SetupInitialInputs()
		} else {
			log.Warn("failed to load config:", sl.Err(err))
			errorMessage := fmt.Sprintf("failed to load config: %s", err.Error())
			panic(errorMessage)
		}
	} else {
		log.Info("config parsed")
		
		log.Warn("attempting to initialize s3 storage")
		s3Storage, err := s3.New(context.Background(), &cfg.S3Config, log)
		if err != nil {
			log.Warn("failed to initialize s3 storage:", sl.Err(err))
			errorMessage := fmt.Sprintf("failed to initialize s3 storage: %s", err.Error())
			panic(errorMessage)
		}
		log.Info("s3 storage initialized")

		log.Warn("download meta data")
		ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
		defer cancel()
		
		meta, err := s3Storage.DownloadMeta(ctx)
		if err != nil {
			log.Warn("failed to download metadata from S3:", sl.Err(err))
			errorMessage := fmt.Sprintf("failed to download metadata from S3: %s", err.Error())
			panic(errorMessage)
		}

		log.Info("meta data downloaded")
		
		initialModel = app.NewModel(s3Storage, cfg, meta, log)
	}

	p := tea.NewProgram(initialModel)

	if _, err = p.Run(); err != nil {
		log.Warn("failed to run program", sl.Err(err))
		errorMessage := fmt.Sprintf("failed to run program: %s", err.Error())
		panic(errorMessage)
	}
		
	if err = file.Truncate(0); err != nil {
		log.Warn("failed to truncate log file: %w", sl.Err(err))
	}

	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}

func setupLogger(w io.Writer) *slog.Logger {
	log := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return log
}
