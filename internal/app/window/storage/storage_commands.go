package storage

type localCommand int

const (
	cmdStorageSetLocations localCommand = iota
	cmdStorageSetSelected
	cmdStorageSetMonitor
	cmdStorageHandleClick
)
