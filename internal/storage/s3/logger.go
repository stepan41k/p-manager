package s3

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/smithy-go/logging"
)

type s3SlogAdapter struct {
	log *slog.Logger
}

func (a s3SlogAdapter) Logf(classification logging.Classification, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	// There is no need to verify the checksum, as it is verified at the encryption level
	if strings.Contains(msg, "supported checksum") {
		return
	}

	a.log.Info(msg, "sdk_class", classification)
}
