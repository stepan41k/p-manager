package exit

import (
	"log/slog"
	"os"
	"os/exec"

	"github.com/stepan41k/p-manager/internal/lib/logger/sl"
)

func GracefullyStop(log *slog.Logger, file *os.File) error {
	if err := file.Truncate(0); err != nil {
		log.Warn("failed to truncate log file: %w", sl.Err(err))
		return err
	}

	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()

	return nil
}
