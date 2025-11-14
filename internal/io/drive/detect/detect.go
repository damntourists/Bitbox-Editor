package detect

import (
	"bitbox-editor/internal/logging"

	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	log = logging.NewLogger("detect")
}
