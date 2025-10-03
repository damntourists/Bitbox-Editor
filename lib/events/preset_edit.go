package events

type PresetEdit int32

const (
	Updated PresetEdit = 0
)

type PresetEditEventRecord struct {
	Type PresetEdit
	Data interface{}
}
