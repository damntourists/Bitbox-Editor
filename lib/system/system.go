package system

import (
	"bitbox-editor/lib/logging"
	"fmt"

	gap "github.com/muesli/go-app-paths"
	"go.uber.org/zap"
)

var log *zap.Logger

var (
	Scope      *gap.Scope
	DataDirs   []string
	CacheDir   string
	ConfigDirs []string
)

func init() {
	log = logging.NewLogger("system")

	Scope = gap.NewScope(gap.User, "bitbox-editor")
	log.Debug(fmt.Sprintf("Scope: %s", Scope.App))

	var err error
	DataDirs, err = Scope.DataDirs()
	if err != nil {
		panic(err)
	}
	log.Debug(fmt.Sprintf("Data Dirs: %s", DataDirs))

	ConfigDirs, err = Scope.ConfigDirs()
	if err != nil {
		panic(err)
	}
	log.Debug(fmt.Sprintf("Config Dirs: %s", ConfigDirs))

	CacheDir, err = Scope.CacheDir()
	if err != nil {
		panic(err)
	}
	log.Debug(fmt.Sprintf("Cache Dir: %s", CacheDir))
}
