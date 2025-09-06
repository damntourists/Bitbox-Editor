package fonts

import "C"
import (
	"bitbox-editor/lib/logging"
	"bitbox-editor/lib/util"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"

	"bitbox-editor/resources"
	"unsafe"

	"github.com/AllenDang/cimgui-go/utils"
)

const (
	lucide         = "fonticons/Lucide/lucide.ttf" //"fonts/lucide.ttf"
	satoshiBlack   = "fonts/Satoshi_Complete/Satoshi-Black.ttf"
	satoshiBold    = "fonts/Satoshi_Complete/Satoshi-Bold.ttf"
	satoshiItalic  = "fonts/Satoshi_Complete/Satoshi-Italic.ttf"
	satoshiLight   = "fonts/Satoshi_Complete/Satoshi-Light.ttf"
	satoshiMedium  = "fonts/Satoshi_Complete/Satoshi-Medium.ttf"
	satoshiRegular = "fonts/Satoshi_Complete/Satoshi-Regular.ttf"

	firaCodeBold     = "fonts/FiraCode/FiraCode-Bold.ttf"
	firaCodeLight    = "fonts/FiraCode/FiraCode-Light.ttf"
	firaCodeMedium   = "fonts/FiraCode/FiraCode-Medium.ttf"
	firaCodeRegular  = "fonts/FiraCode/FiraCode-Regular.ttf"
	firaCodeRetina   = "fonts/FiraCode/FiraCode-Retina.ttf"
	firaCodeSemiBold = "fonts/FiraCode/FiraCode-SemiBold.ttf"
)

var (
	//globalFontScale = float32(ebiten.Monitor().DeviceScaleFactor())
	globalFontScale = float32(1)
	pointSize       = 1
	fontSizeH1      = 32 * (float32(pointSize) * globalFontScale)
	fontSizeH2      = 24 * (float32(pointSize) * globalFontScale)
	fontSizeH3      = 19 * (float32(pointSize) * globalFontScale)
	fontSizeH4      = 16 * (float32(pointSize) * globalFontScale)
)

var (
	FontDefault *imgui.Font
	FontRegular *imgui.Font
	//FontDefaultItalic *imgui.Font
	FontDefaultBlack *imgui.Font
	FontCode         *imgui.Font
	FontIconsH1      *imgui.Font
	FontIconsH2      *imgui.Font
	FontIconsH3      *imgui.Font
	FontIconsH4      *imgui.Font
	FontsInitialized bool
)

// Keep references to font data while building the atlas to prevent GC
var fontDataKeep [][]byte

func RebuildFonts(backend backend.Backend[glfwbackend.GLFWWindowFlags]) {
	FontsInitialized = false
	// ... existing code ...

	// Clear out any existing fonts
	fontAtlas := imgui.CurrentIO().Fonts()
	if fontAtlas.FontCount() > 0 {
		fontAtlas.Clear()
		logging.Logger.Debug("Cleared existing font atlas.")
	}

	// Reset kept data before rebuilding
	fontDataKeep = nil

	FontDefault = CreateFont(fontAtlas, satoshiRegular, fontSizeH4, false, false)
	FontRegular = CreateFont(fontAtlas, satoshiRegular, fontSizeH4, true, true)
	//FontDefaultItalic = CreateFont(fontAtlas, satoshiItalic, fontSizeH4, true)
	FontDefaultBlack = CreateFont(fontAtlas, satoshiBlack, fontSizeH4, true, false)
	//FontIconsH1 = CreateFont(fontAtlas, lucide, fontSizeH1, true)
	//FontIconsH2 = CreateFont(fontAtlas, lucide, fontSizeH2, true)
	//FontIconsH3 = CreateFont(fontAtlas, lucide, fontSizeH3, true)
	FontIconsH4 = CreateFont(fontAtlas, lucide, fontSizeH4, true, false)
	FontCode = CreateFont(fontAtlas, firaCodeRegular, fontSizeH4, false, false)

	// Force font atlas to rebuild tex cache
	//_, _, _, _ = fontAtlas.GetTextureDataAsRGBA32()

	fontTextureImg, w, h, _ := fontAtlas.GetTextureDataAsRGBA32()
	tex := backend.CreateTexture(fontTextureImg, int(w), int(h))

	imgui.CurrentIO().Fonts().SetTexID(tex)
	imgui.CurrentIO().Fonts().SetTexReady(true)

	// At this point the atlas/texture is ready; we can release our Go font data
	fontDataKeep = nil

	FontsInitialized = true
}

func CreateFont(atlas *imgui.FontAtlas, path string, size float32, addicons bool, makeDefault bool) *imgui.Font {
	data, err := resources.Assets.ReadFile(path)
	util.PanicOnError(err)

	// Keep a reference so the GC doesn't move/free it while ImGui reads it
	fontDataKeep = append(fontDataKeep, data)

	dataPtr := uintptr(unsafe.Pointer(utils.SliceToPtr(data)))
	dataLen := int32(len(data))

	var font *imgui.Font
	if !addicons {
		cfg := imgui.NewFontConfig()
		ranges := atlas.GlyphRangesDefault()
		// IMPORTANT: Do NOT let atlas free Go-owned memory
		cfg.SetFontDataOwnedByAtlas(false)
		font = atlas.AddFontFromMemoryTTFV(dataPtr, dataLen, size, cfg, ranges)
	} else {
		cfg := imgui.NewFontConfig()
		cfg.SetMergeMode(true) // Merges icons into the previous font.
		cfg.SetGlyphOffset(imgui.Vec2{X: 0, Y: 3})
		//cfg.SetRasterizerMultiply(2)
		//cfg.SetGlyphMinAdvanceX(fontSizeH1) // Use to make icon monospaced
		cfg.SetPixelSnapH(true)
		cfg.SetOversampleH(2)
		cfg.SetOversampleV(2)

		// IMPORTANT: Do NOT let atlas free Go-owned memory
		cfg.SetFontDataOwnedByAtlas(false)

		ranges := imgui.NewGlyphRange()
		rangesWChar := []imgui.Wchar{
			imgui.Wchar(icons.Min),
			imgui.Wchar(icons.Max16),
			0,
		}
		builder := imgui.NewFontGlyphRangesBuilder()
		// TODO Add all known icons that are used here.
		for _, txt := range IconsInUse {
			builder.AddText(txt)
		}
		builder.AddRanges(utils.SliceToPtr(rangesWChar))
		builder.BuildRanges(ranges)

		font = atlas.AddFontFromMemoryTTFV(dataPtr, dataLen, size, cfg, ranges.Data())
		if makeDefault {
			imgui.CurrentIO().Fonts().AddFontDefaultV(cfg)
		}
	}

	return font
}
