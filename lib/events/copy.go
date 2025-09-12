package events

type FileCopy int32

const (
	FileCopyCompletedEvent FileCopy = 0
	FileCopyFailedEvent    FileCopy = 1
	FileCopyProgressEvent  FileCopy = 2
)

type FileCopyEventRecord struct {
	Type        FileCopy
	Progress    float64
	Completed   bool
	Failed      bool
	Source      string
	Destination string
}
