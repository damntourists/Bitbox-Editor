package table

type SortFunc func(SortDirection)

type SortDirection byte

const (
	SortAscending  SortDirection = 1
	SortDescending SortDirection = 2
)
