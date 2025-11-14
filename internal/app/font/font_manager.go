package font

import (
	"bitbox-editor/internal/logging"
	"bitbox-editor/resources"
	"fmt"
	"unsafe"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/utils"
	"go.uber.org/zap"
)

var log = logging.NewLogger("font")

// FontManager manages all fonts in the application
type FontManager struct {
	globalScale    float32
	fonts          map[string]*FontBuilder
	backend        backend.Backend[glfwbackend.GLFWWindowFlags]
	fontDataKeep   [][]byte
	iconRanges     imgui.GlyphRange
	iconRangesData []imgui.Wchar
}

var manager *FontManager

// Init initializes the font manager
func Init(backend backend.Backend[glfwbackend.GLFWWindowFlags]) {
	manager = &FontManager{
		globalScale: 1.0,
		fonts:       make(map[string]*FontBuilder),
		backend:     backend,
	}
}

// Manager returns the global font manager instance
func Manager() *FontManager {
	if manager == nil {
		panic("FontManager not initialized. Call font.Init() first.")
	}
	return manager
}

// SetGlobalScale sets the global font scale
func SetGlobalScale(scale float32) {
	Manager().globalScale = scale
}

// GetGlobalScale returns the current global scale
func GetGlobalScale() float32 {
	return Manager().globalScale
}

// RebuildFonts rebuilds all fonts with current settings
func RebuildFonts() {
	Manager().rebuild()
}

// NewFont returns a font builder for the specified font family
func NewFont(family FontFamily) *FontBuilder {
	// Check if we have a cached builder for this family
	if builder, exists := Manager().fonts[string(family)]; exists {
		return builder.copy()
	}

	// Create new builder with defaults
	builder := &FontBuilder{
		family:       family,
		size:         14.0,
		weight:       FontWeightRegular,
		includeIcons: false,
		glyphOffset:  imgui.Vec2{X: 0, Y: 3},
		oversampleH:  2,
		oversampleV:  2,
	}

	return builder
}

// rebuild clears and rebuilds all fonts
func (m *FontManager) rebuild() {
	FontsInitialized = false

	fontAtlas := imgui.CurrentIO().Fonts()
	if fontAtlas.FontCount() > 0 {
		fontAtlas.Clear()
		log.Debug("Cleared existing font atlas.")
	}

	m.fontDataKeep = nil
	m.buildIconRanges()

	fontOrder := []string{
		"default",
		"regular",
		"black",
		"code",
		"pacifico",
		"hegarty",
		"bebas",
		"nabla",
		"icons_h4",
	}

	for _, name := range fontOrder {
		builder, exists := m.fonts[name]
		if !exists {
			log.Warn("Font not registered", zap.String("name", name))
			continue
		}

		font := builder.build(fontAtlas, m)
		log.Debug(fmt.Sprintf("Built font %s", name))

		switch name {
		case "default":
			FontDefault = font
		case "regular":
			FontRegular = font
		case "black":
			FontDefaultBlack = font
		case "code":
			FontCode = font
		case "pacifico":
			FontPacifico = font
		case "hegarty":
			FontHegarty = font
		case "bebas":
			FontBebas = font
		case "nabla":
			FontNabla = font
		case "icons_h4":
			FontIconsH4 = font
		}
	}

	// Release font data
	m.fontDataKeep = nil

	FontsInitialized = true
	log.Debug("Fonts rebuilt successfully", zap.Int("fontCount", int(fontAtlas.FontCount())))
}

// buildIconRanges builds the glyph ranges for icons
func (m *FontManager) buildIconRanges() {
	// Build ranges from icons in use
	rangesWChar := []imgui.Wchar{
		imgui.Wchar(icons.Min),
		imgui.Wchar(icons.Max16),
		0,
	}

	builder := imgui.NewFontGlyphRangesBuilder()
	for _, txt := range IconsInUse {
		builder.AddText(txt)
	}
	builder.AddRanges(utils.SliceToPtr(rangesWChar))

	ranges := imgui.NewGlyphRange()
	builder.BuildRanges(ranges)
	m.iconRanges = ranges
	m.iconRangesData = rangesWChar
}

// Register registers a font builder with a name
func (m *FontManager) Register(name string, builder *FontBuilder) {
	m.fonts[name] = builder
}

// loadFontData loads font data from resources and keeps it in memory
func (m *FontManager) loadFontData(path string) (uintptr, int32, error) {
	data, err := resources.Assets.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}

	// Keep the data in memory
	m.fontDataKeep = append(m.fontDataKeep, data)

	dataPtr := uintptr(unsafe.Pointer(utils.SliceToPtr(data)))
	dataLen := int32(len(data))

	return dataPtr, dataLen, nil
}
