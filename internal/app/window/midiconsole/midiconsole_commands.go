package midiconsole

type localCommand int

const (
	cmdMidiPorts localCommand = iota
	cmdMidiPortSelected
	cmdMidiPortMonitor
	cmdMidiPortSelectionChanged
)
