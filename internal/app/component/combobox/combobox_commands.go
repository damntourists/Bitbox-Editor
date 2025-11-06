package combobox

type localCommand int

// These are private to the combobox component.
const (
	cmdSetComboBoxItems localCommand = iota
	cmdSetComboBoxSelected
	cmdSetComboBoxPreview
	cmdSetComboBoxFlags
)
