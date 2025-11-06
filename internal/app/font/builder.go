package font

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

// FontBuilder fluent API for building fonts
type FontBuilder struct {
	family       FontFamily
	size         float32
	weight       FontWeight
	includeIcons bool
	glyphOffset  imgui.Vec2
	oversampleH  int
	oversampleV  int
	pixelSnapH   bool
	mergeMode    bool
	builtFont    *imgui.Font
}

// copy creates a copy of the builder for chaining
func (b *FontBuilder) copy() *FontBuilder {
	return &FontBuilder{
		family:       b.family,
		size:         b.size,
		weight:       b.weight,
		includeIcons: b.includeIcons,
		glyphOffset:  b.glyphOffset,
		oversampleH:  b.oversampleH,
		oversampleV:  b.oversampleV,
		pixelSnapH:   b.pixelSnapH,
		mergeMode:    b.mergeMode,
		builtFont:    b.builtFont,
	}
}

// SetSize sets the font size in pixels (scaled by the global scale)
func (b *FontBuilder) SetSize(size float32) *FontBuilder {
	b.size = size
	return b
}

// SetWeight sets the font weight
func (b *FontBuilder) SetWeight(weight FontWeight) *FontBuilder {
	b.weight = weight
	return b
}

// IncludeIcons merges the lucide icons font with this font
func (b *FontBuilder) IncludeIcons(include bool) *FontBuilder {
	b.includeIcons = include
	return b
}

// SetGlyphOffset sets the vertical offset for glyphs to adjust alignment
func (b *FontBuilder) SetGlyphOffset(offset imgui.Vec2) *FontBuilder {
	b.glyphOffset = offset
	return b
}

// SetOversample sets the oversampling values for better quality
func (b *FontBuilder) SetOversample(h, v int) *FontBuilder {
	b.oversampleH = h
	b.oversampleV = v
	return b
}

// SetPixelSnapH enables pixel snapping for horizontal alignment
func (b *FontBuilder) SetPixelSnapH(snap bool) *FontBuilder {
	b.pixelSnapH = snap
	return b
}

// Build the font and return the imgui.Font pointer
func (b *FontBuilder) Build() *imgui.Font {
	if !FontsInitialized {
		// Fonts not ready, return nil
		return nil
	}
	return b.builtFont
}

// build is the internal build method called by the manager
func (b *FontBuilder) build(atlas *imgui.FontAtlas, manager *FontManager) *imgui.Font {
	// Get the font path based on family and weight
	fontPath := b.getFontPath()

	// Apply global scale to size
	scaledSize := b.size * manager.globalScale

	// Load the main font data
	dataPtr, dataLen, err := manager.loadFontData(fontPath)
	if err != nil {
		panic("Failed to load font: " + fontPath + " - " + err.Error())
	}

	// Create font config
	cfg := imgui.NewFontConfig()
	cfg.SetFontDataOwnedByAtlas(false)
	cfg.SetOversampleH(b.oversampleH)
	cfg.SetOversampleV(b.oversampleV)
	cfg.SetPixelSnapH(b.pixelSnapH)

	// Add the main font
	ranges := atlas.GlyphRangesDefault()
	font := atlas.AddFontFromMemoryTTFV(dataPtr, dataLen, scaledSize, cfg, ranges)

	if font == nil {
		panic("Failed to create font from " + fontPath)
	}

	// If icons should be included, merge them into this font
	if b.includeIcons {
		b.mergeIcons(atlas, manager, scaledSize)
	}

	b.builtFont = font
	return font
}

// mergeIcons merges the lucide icon font into the current font
func (b *FontBuilder) mergeIcons(atlas *imgui.FontAtlas, manager *FontManager, size float32) {
	// Load icon font data
	iconPath := fontPaths[FontFamilyLucide][FontWeightRegular]
	dataPtr, dataLen, err := manager.loadFontData(iconPath)
	if err != nil {
		panic("Failed to load icon font: " + iconPath + " - " + err.Error())
	}

	// Create merge config
	cfg := imgui.NewFontConfig()
	cfg.SetMergeMode(true)
	cfg.SetFontDataOwnedByAtlas(false)
	cfg.SetGlyphOffset(b.glyphOffset)
	cfg.SetPixelSnapH(b.pixelSnapH)
	cfg.SetOversampleH(b.oversampleH)
	cfg.SetOversampleV(b.oversampleV)

	// Add icon font with merge mode
	mergedFont := atlas.AddFontFromMemoryTTFV(dataPtr, dataLen, size, cfg, manager.iconRanges.Data())
	if mergedFont == nil {
		panic("Failed to merge icon font into atlas")
	}
}

// getFontPath returns the resource path for the font based on family and weight
func (b *FontBuilder) getFontPath() string {
	if paths, ok := fontPaths[b.family]; ok {
		if path, ok := paths[b.weight]; ok {
			return path
		}
	}

	// Fallback to regular weight if requested weight doesn't exist
	if paths, ok := fontPaths[b.family]; ok {
		if path, ok := paths[FontWeightRegular]; ok {

			return path
		}
	}

	// Final fallback
	return fontPaths[FontFamilySatoshi][FontWeightRegular]
}
