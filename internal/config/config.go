package config

import (
	"bitbox-editor/internal/logging"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	APP_NAME = "bitbox-editor"
)

// defaultConfig is written if a configuration file is absent from the system.
var defaultConfig = []byte(`
# Default pre-generated config

[global]
loglevel = "DEBUG"
theme = "future dark"
console_max_lines = 100
colormap = "Cool"

[toolbar]
placement = "top" # Options: "top", "left", "right", "bottom"
padding = 4          
size = 48            
margin = 0           

[dockspace]
window_padding = 0   

[spectrum]
attack_time = 5
attack_easing = "EaseOutCubic"
decay_time = 50
decay_easing = "EaseLinear"
peak_hold_time = 150
peak_fall_speed = 0.020
noise_gate = 0.05
color_mode = "height"
static_color_idx = 128
`)

var log *zap.Logger

func init() {
	log = logging.NewLogger("config")

	configdir, _ := os.UserConfigDir()
	appConfigDir := filepath.Join(configdir, APP_NAME)

	viper.AutomaticEnv()
	viper.SetEnvPrefix(APP_NAME)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(appConfigDir)

	// Set all defaults so they're preserved during WriteConfig
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError

		if errors.As(err, &configFileNotFoundError) {
			log.Debug("Config file missing!")

			err = viper.ReadConfig(bytes.NewBuffer(defaultConfig))
			if err != nil {
				panic(err)
			}

			// ensure config directory exists; create it if not ...
			_, err := os.Open(appConfigDir)
			if err != nil {
				// Missing, create it.
				log.Debug(fmt.Sprintf("Creating directory: %s ...", appConfigDir))
				mkdirerr := os.Mkdir(appConfigDir, 0750)
				if mkdirerr != nil {
					panic(mkdirerr)
				}
			}

			err = viper.SafeWriteConfig()
			if err != nil {
				panic(err)
			}
			log.Debug("Wrote default config successfully.")

			if err := viper.ReadInConfig(); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		log.Debug("Config file found!")
	}

	// Watch config file for changes and automatically reload
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debug("Config file changed, reloading...", zap.String("file", e.Name))
	})
}

/*
╔════════╗
║ Config ║
╚════════╝
*/

// Config structure representing the application configuration
type Config struct {
	Global struct {
		Loglevel        string
		Theme           string
		ConsoleMaxLines int
		Colormap        string
	}

	Toolbar struct {
		Placement string
		Padding   float32
		Size      float32
		Margin    float32
	}

	Dockspace struct {
		WindowPadding float32
	}

	Spectrum struct {
		AttackTime     int
		AttackEasing   string
		DecayTime      int
		DecayEasing    string
		PeakHoldTime   int
		PeakFallSpeed  float64
		NoiseGate      float64
		ColorMode      string
		StaticColorIdx int
	}
}

// setDefaults sets all default values in viper
func setDefaults() {
	// Global defaults
	viper.SetDefault("global.loglevel", "DEBUG")
	viper.SetDefault("global.theme", "future dark")
	viper.SetDefault("global.console_max_lines", 100)
	viper.SetDefault("global.colormap", "Cool")

	// Toolbar defaults
	viper.SetDefault("toolbar.placement", "top")
	viper.SetDefault("toolbar.padding", 4.0)
	viper.SetDefault("toolbar.size", 48.0)
	viper.SetDefault("toolbar.margin", 0.0)

	// Dockspace defaults
	viper.SetDefault("dockspace.window_padding", 0.0)

	// Spectrum defaults
	viper.SetDefault("spectrum.attack_time", 5)
	viper.SetDefault("spectrum.attack_easing", "EaseOutCubic")
	viper.SetDefault("spectrum.decay_time", 50)
	viper.SetDefault("spectrum.decay_easing", "EaseLinear")
	viper.SetDefault("spectrum.peak_hold_time", 150)
	viper.SetDefault("spectrum.peak_fall_speed", 0.020)
	viper.SetDefault("spectrum.noise_gate", 0.05)
	viper.SetDefault("spectrum.color_mode", "height")
	viper.SetDefault("spectrum.static_color_idx", 128)
}

/*
╭──────────────╮
│ Theme Config │
╰──────────────╯
*/

// SetTheme updates the theme in the config and saves it
func SetTheme(themeName string) error {
	viper.Set("global.theme", themeName)
	err := viper.WriteConfig()
	if err != nil {
		log.Error("Failed to save theme to config", zap.String("theme", themeName), zap.Error(err))
		return err
	}
	log.Debug("Saved theme to config", zap.String("theme", themeName))
	return nil
}

// GetTheme retrieves the theme name from config
func GetTheme() string {
	return viper.GetString("global.theme")
}

// SetColormap updates the colormap in the config and saves it
func SetColormap(colormap string) error {
	viper.Set("global.colormap", colormap)
	err := viper.WriteConfig()
	if err != nil {
		log.Error("Failed to save colormap to config", zap.String("colormap", colormap), zap.Error(err))
		return err
	}
	log.Debug("Saved colormap to config", zap.String("colormap", colormap))
	return nil
}

// GetColormap retrieves the colormap from config
func GetColormap() string {
	colormap := viper.GetString("global.colormap")
	if colormap == "" {
		return "Cool" // Default fallback
	}
	return colormap
}

/*
╭────────────────╮
│ Console Config │
╰────────────────╯
*/

// SetConsoleMaxLines updates the console max lines in the config and saves it
func SetConsoleMaxLines(maxLines int) error {
	viper.Set("global.console_max_lines", maxLines)
	err := viper.WriteConfig()
	if err != nil {
		log.Error("Failed to save console max lines to config", zap.Int("maxLines", maxLines), zap.Error(err))
		return err
	}
	log.Debug("Saved console max lines to config", zap.Int("maxLines", maxLines))
	return nil
}

// GetConsoleMaxLines retrieves the console max lines from config
func GetConsoleMaxLines() int {
	maxLines := viper.GetInt("global.console_max_lines")
	if maxLines <= 0 {
		return 100 // Default fallback
	}
	return maxLines
}

/*
╭────────────────╮
│ Toolbar Config │
╰────────────────╯
*/

// SetToolbarPlacement updates the toolbar placement in config
func SetToolbarPlacement(placement string) error {
	viper.Set("toolbar.placement", placement)
	return viper.WriteConfig()
}

// GetToolbarPlacement retrieves the toolbar placement from config
func GetToolbarPlacement() string {
	placement := viper.GetString("toolbar.placement")
	if placement == "" {
		return "top"
	}
	return placement
}

// SetToolbarPadding updates the toolbar padding in config
func SetToolbarPadding(padding float32) error {
	viper.Set("toolbar.padding", padding)
	return viper.WriteConfig()
}

// GetToolbarPadding retrieves the toolbar padding from config
func GetToolbarPadding() float32 {
	padding := viper.GetFloat64("toolbar.padding")
	if padding < 0 {
		return 4.0
	}
	return float32(padding)
}

// SetToolbarSize updates the toolbar size in config
func SetToolbarSize(size float32) error {
	viper.Set("toolbar.size", size)
	return viper.WriteConfig()
}

// GetToolbarSize retrieves the toolbar size from config
func GetToolbarSize() float32 {
	size := viper.GetFloat64("toolbar.size")
	if size <= 0 {
		return 48.0
	}
	return float32(size)
}

// SetToolbarMargin updates the toolbar margin in config
func SetToolbarMargin(margin float32) error {
	viper.Set("toolbar.margin", margin)
	return viper.WriteConfig()
}

// GetToolbarMargin retrieves the toolbar margin from config
func GetToolbarMargin() float32 {
	margin := viper.GetFloat64("toolbar.margin")
	if margin < 0 {
		return 0.0
	}
	return float32(margin)
}

/*
╭─────────╮
│ Docking │
╰─────────╯
*/

// SetDockspaceWindowPadding updates the dockspace window padding in config
func SetDockspaceWindowPadding(padding float32) error {
	viper.Set("dockspace.window_padding", padding)
	return viper.WriteConfig()
}

// GetDockspaceWindowPadding retrieves the dockspace window padding from config
func GetDockspaceWindowPadding() float32 {
	padding := viper.GetFloat64("dockspace.window_padding")
	if padding < 0 {
		return 0.0
	}
	return float32(padding)
}

/*
╭──────────╮
│ Spectrum │
╰──────────╯
*/

// SetSpectrumAttackTime updates the spectrum attack time in config
func SetSpectrumAttackTime(attackTime int) error {
	viper.Set("spectrum.attack_time", attackTime)
	return viper.WriteConfig()
}

// GetSpectrumAttackTime retrieves the spectrum attack time from config
func GetSpectrumAttackTime() int {
	attackTime := viper.GetInt("spectrum.attack_time")
	if attackTime <= 0 {
		return 5
	}
	return attackTime
}

// SetSpectrumAttackEasing updates the spectrum attack easing in config
func SetSpectrumAttackEasing(easing string) error {
	viper.Set("spectrum.attack_easing", easing)
	return viper.WriteConfig()
}

// GetSpectrumAttackEasing retrieves the spectrum attack easing from config
func GetSpectrumAttackEasing() string {
	easing := viper.GetString("spectrum.attack_easing")
	if easing == "" {
		return "EaseOutCubic"
	}
	return easing
}

// SetSpectrumDecayTime updates the spectrum decay time in config
func SetSpectrumDecayTime(decayTime int) error {
	viper.Set("spectrum.decay_time", decayTime)
	return viper.WriteConfig()
}

// GetSpectrumDecayTime retrieves the spectrum decay time from config
func GetSpectrumDecayTime() int {
	decayTime := viper.GetInt("spectrum.decay_time")
	if decayTime <= 0 {
		return 50
	}
	return decayTime
}

// SetSpectrumDecayEasing updates the spectrum decay easing in config
func SetSpectrumDecayEasing(easing string) error {
	viper.Set("spectrum.decay_easing", easing)
	return viper.WriteConfig()
}

// GetSpectrumDecayEasing retrieves the spectrum decay easing from config
func GetSpectrumDecayEasing() string {
	easing := viper.GetString("spectrum.decay_easing")
	if easing == "" {
		return "EaseLinear"
	}
	return easing
}

// SetSpectrumPeakHoldTime updates the spectrum peak hold time in config
func SetSpectrumPeakHoldTime(peakHoldTime int) error {
	viper.Set("spectrum.peak_hold_time", peakHoldTime)
	return viper.WriteConfig()
}

// GetSpectrumPeakHoldTime retrieves the spectrum peak hold time from config
func GetSpectrumPeakHoldTime() int {
	peakHoldTime := viper.GetInt("spectrum.peak_hold_time")
	if peakHoldTime <= 0 {
		return 150
	}
	return peakHoldTime
}

// SetSpectrumPeakFallSpeed updates the spectrum peak fall speed in config
func SetSpectrumPeakFallSpeed(peakFallSpeed float64) error {
	viper.Set("spectrum.peak_fall_speed", peakFallSpeed)
	return viper.WriteConfig()
}

// GetSpectrumPeakFallSpeed retrieves the spectrum peak fall speed from config
func GetSpectrumPeakFallSpeed() float64 {
	peakFallSpeed := viper.GetFloat64("spectrum.peak_fall_speed")
	if peakFallSpeed <= 0 {
		return 0.020
	}
	return peakFallSpeed
}

// SetSpectrumNoiseGate updates the spectrum noise gate in config
func SetSpectrumNoiseGate(noiseGate float64) error {
	viper.Set("spectrum.noise_gate", noiseGate)
	return viper.WriteConfig()
}

// GetSpectrumNoiseGate retrieves the spectrum noise gate from config
func GetSpectrumNoiseGate() float64 {
	noiseGate := viper.GetFloat64("spectrum.noise_gate")
	if noiseGate < 0 {
		return 0.05
	}
	return noiseGate
}

// SetSpectrumColorMode updates the spectrum color mode in config
func SetSpectrumColorMode(colorMode string) error {
	viper.Set("spectrum.color_mode", colorMode)
	return viper.WriteConfig()
}

// GetSpectrumColorMode retrieves the spectrum color mode from config
func GetSpectrumColorMode() string {
	colorMode := viper.GetString("spectrum.color_mode")
	if colorMode == "" {
		return "height"
	}
	return colorMode
}

// SetSpectrumStaticColorIdx updates the spectrum static color index in config
func SetSpectrumStaticColorIdx(staticColorIdx int) error {
	viper.Set("spectrum.static_color_idx", staticColorIdx)
	return viper.WriteConfig()
}

// GetSpectrumStaticColorIdx retrieves the spectrum static color index from config
func GetSpectrumStaticColorIdx() int {
	staticColorIdx := viper.GetInt("spectrum.static_color_idx")
	if staticColorIdx < 0 {
		return 128
	}
	return staticColorIdx
}
