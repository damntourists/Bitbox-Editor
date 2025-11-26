package events

const (
	PadGridSelectKey = "padgrid.select"
)

// PadGridSelectEvent is published when a pad is selected in a pad grid.
type PadGridSelectEvent struct {
	Pad     interface{}
	OwnerID string
}

func (e PadGridSelectEvent) Type() string {
	return PadGridSelectKey
}

func (e PadGridSelectEvent) GetOwnerID() string {
	return e.OwnerID
}
