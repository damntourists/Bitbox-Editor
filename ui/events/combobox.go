package events

type ComboBoxEventType int32

const (
	SelectionChanged ComboBoxEventType = 0
)

type ComboBoxEventRecord struct {
	Type ComboBoxEventType
	Data interface{}
}
