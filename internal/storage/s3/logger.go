package s3

import (
    "github.com/aws/smithy-go/logging"
    "log/slog"
    "fmt"
)

type s3SlogAdapter struct {
	log *slog.Logger
}

func (a s3SlogAdapter) Logf(classification logging.Classification, format string, v ...interface{}) {
    msg := fmt.Sprintf(format, v...)
    a.log.Info(msg, "sdk_class", classification)
}