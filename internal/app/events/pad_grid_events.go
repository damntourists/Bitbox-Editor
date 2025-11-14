package events

type PadGridEvent int32

const (
	PadGridSelectEvent PadGridEvent = iota
)
const (
	PadGridSelectKey = "padgrid.select"
)

type PadGridEventRecord struct {
	EventType PadGridEvent
	Pad       interface{}
	// OwnerID is the UUID of the window that owns this pad grid
	OwnerID string
}

func (e PadGridEventRecord) Type() string {
	switch e.EventType {
	case PadGridSelectEvent:
		return PadGridSelectKey
	default:
		return "padgrid.unknown"
	}
}

func (e PadGridEventRecord) GetOwnerID() string {
	return e.OwnerID
}
