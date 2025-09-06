package copy

import (
	"bitbox-editor/lib/events"
	"bitbox-editor/lib/logging"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// WriteCounter counts the number of bytes written and prints progress.
type WriteCounter struct {
	Total    uint64
	FileSize int64

	src   string
	dest  string
	group string
}

// Write implements the io.Writer interface.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.EmitProgress()
	return n, nil
}

// EmitProgress emits the current progress as a signal.
func (wc *WriteCounter) EmitProgress() {
	percentage := float64(wc.Total) / float64(wc.FileSize) * 100
	//Logger.Debug(
	//	fmt.Sprintf("\r%d bytes copied / %d total bytes (%.2f%%)",
	//		wc.Total,
	//		wc.FileSize,
	//		percentage,
	//	),
	//)
	events.FileCopyEvent.Emit(
		context.Background(),
		events.FileCopyEventRecord{
			Type:        events.FileCopyProgressEvent,
			Source:      wc.src,
			Destination: wc.dest,
			Progress:    percentage,
			Completed:   false,
			Failed:      false,
		},
	)
}

func emitCopyFailed(src string, dest string) {
	events.FileCopyEvent.Emit(
		context.Background(),
		events.FileCopyEventRecord{
			Type:        events.FileCopyFailedEvent,
			Source:      src,
			Destination: dest,
			Progress:    0,
			Completed:   false,
			Failed:      true,
		},
	)
}

func emitCopyCompleted(src string, dest string) {
	events.FileCopyEvent.Emit(
		context.Background(),
		events.FileCopyEventRecord{
			Type:        events.FileCopyCompletedEvent,
			Source:      src,
			Destination: dest,
			Progress:    100,
			Completed:   true,
			Failed:      false,
		},
	)
}

// CopyFileWithProgress copies a file and displays the progress.
func CopyFileWithProgress(src, dest string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		emitCopyFailed(src, dest)
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		emitCopyFailed(src, dest)
		return errors.New(fmt.Sprintf("%s is not a regular file", src))
	}

	source, err := os.Open(src)
	if err != nil {
		emitCopyFailed(src, dest)
		return err
	}
	defer source.Close()

	destination, err := os.Create(dest)
	if err != nil {
		emitCopyFailed(src, dest)
		return err
	}

	defer destination.Close()

	counter := &WriteCounter{
		FileSize: sourceFileStat.Size(),
		src:      src,
		dest:     dest,
	}
	_, err = io.Copy(destination, io.TeeReader(source, counter))

	emitCopyCompleted(src, dest)
	return err
}

func GetFileChecksum(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return nil, err
	}

	return hasher.Sum(nil), nil
}

func GetFileNameWithoutExtension(fileName string) string {
	logging.Logger.Debug(fmt.Sprintf("GetFileNameWithoutExtension: %s", fileName))
	return strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
}
