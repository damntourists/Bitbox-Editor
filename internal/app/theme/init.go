package theme

import (
	"bitbox-editor/internal/config"
	"bitbox-editor/internal/logging"
	"bitbox-editor/internal/util"
	"embed"

	"github.com/AllenDang/cimgui-go/implot"
	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

const DEFAULT_THEME = "future dark"
const DEFAULT_COLORMAP = "jet"

var (
	Themes          ThemeCollection
	currentTheme    *Theme
	currentColorMap implot.Colormap
)

//go:embed themes.toml
var EmbededThemes embed.FS

func init() {
	log := logging.NewLogger("theme")

	collection := ThemeCollection{Themes: make([]*Theme, 0)}

	data, err := EmbededThemes.ReadFile("themes.toml")
	if err != nil {
		return
	}
	util.PanicOnError(err)

	_, err = toml.Decode(string(data), &collection)
	util.PanicOnError(err)

	Themes = collection

	// Load colormap from config
	colormapName := config.GetColormap()
	currentColorMap = GetColormapByName(colormapName)

	// Load theme from config, fallback to default if not found
	themeName := config.GetTheme()
	if themeName == "" {
		themeName = DEFAULT_THEME
	}

	theme, err := GetThemeByName(themeName)
	if err != nil {
		log.Warn("Failed to load theme from config, using default",
			zap.String("configTheme", themeName),
			zap.String("defaultTheme", DEFAULT_THEME),
			zap.Error(err))
		theme, _ = GetThemeByName(DEFAULT_THEME)
	} else {
		log.Debug("Loaded theme from config", zap.String("theme", themeName))
	}

	currentTheme = theme
}
