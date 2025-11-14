package table

type localCommand int

const (
	// TableComponent commands
	cmdSetTableFlags localCommand = iota
	cmdSetTableInnerWidth
	cmdSetTableRows
	cmdSetTableColumns
	cmdSetTableFastMode
	cmdSetTableFreeze
	cmdSetTableNoHeader

	// TableColumnComponent commands
	cmdSetTableColumnFlags
	cmdSetTableColumnWidthOrWeight
	cmdSetTableColumnUserID
	cmdSetTableColumnSortFn

	// TableRowComponent commands
	cmdSetTableRowFlags
	cmdSetTableRowMinHeight
	cmdSetTableRowLayout
)

type TableFreezePayload struct {
	Col int
	Row int
}
