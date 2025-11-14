package text

type localCommand int

const (
	cmdSetTextFont localCommand = iota
	cmdSetTextWrapped
	cmdSetTextSelectable
)
