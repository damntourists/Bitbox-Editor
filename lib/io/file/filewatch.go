package file

import (
	"bitbox-editor/lib/logging"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

var FileWatch *fsnotify.Watcher
var log *zap.Logger

type FileWatchEventFunc func(event fsnotify.Event)

func init() {
	log = logging.NewLogger("filewatch")
}

func WatchDir(dir string, function FileWatchEventFunc) {
	FileWatch, _ = fsnotify.NewWatcher()
	defer FileWatch.Close()

	log.Debug(fmt.Sprintf("Adding file watcher to: %s", dir))
	err := filepath.Walk(dir, watchDir)

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-FileWatch.Events:
				function(event)
			case err = <-FileWatch.Errors:
				log.Debug(fmt.Sprintf("[channel err]-> %s", err.Error()))
			}
		}
	}()

	<-done
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func watchDir(path string, fi os.FileInfo, _ error) error {
	if fi.Mode().IsDir() {
		return FileWatch.Add(path)
	}
	return nil
}
