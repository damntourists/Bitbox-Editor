package library

type localCommand int

const (
	cmdLibSetStorageLoc localCommand = iota
	cmdLibSetScanning
	cmdLibSetFSTree
	cmdLibSetTreeRows
	cmdLibSetSearchQuery
)
