package tree

type localCommand int

const (
	cmdSetTreeFlags localCommand = iota
	cmdSetTreeColumns
	cmdSetTreeRows
	cmdSetTreeFreeze
)
