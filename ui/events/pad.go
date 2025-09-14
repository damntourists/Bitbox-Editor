package events

type PadEventType int32

const (
	PadActivated PadEventType = 0
)

type PadEventRecord struct {
	Type PadEventType
	Data interface{}
}
