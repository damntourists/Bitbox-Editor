package settings

type localCommand int

const (
	cmdSettingsSetTheme localCommand = iota
	cmdSettingsUpdateCurrentThemeName
	cmdSettingsSetConsoleMaxLines
	cmdSettingsSetColormap
	cmdSettingsSetSpectrumSettings
)
