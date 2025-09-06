package detect

import (
	"bitbox-editor/lib/logging"

	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	log = logging.NewLogger("detect")
}
