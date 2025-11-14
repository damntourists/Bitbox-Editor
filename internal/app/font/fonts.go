package font

import (
	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
)

// TODO: Clean this up.

var (
	FontDefault      *imgui.Font
	FontRegular      *imgui.Font
	FontDefaultBlack *imgui.Font
	FontCode         *imgui.Font
	FontIconsH1      *imgui.Font
	FontIconsH2      *imgui.Font
	FontIconsH3      *imgui.Font
	FontIconsH4      *imgui.Font

	FontBebas        *imgui.Font
	FontPacifico     *imgui.Font
	FontHegarty      *imgui.Font
	FontNabla        *imgui.Font
	FontsInitialized bool
)

// InitAndRebuildFonts initializes the font manager and rebuilds all fonts
func InitAndRebuildFonts(backend backend.Backend[glfwbackend.GLFWWindowFlags]) {
	if manager == nil {
		Init(backend)
	}

	registerDefaultFonts()

	manager.rebuild()
}

// registerDefaultFonts registers the default fonts used in the application
func registerDefaultFonts() {
	manager.fonts = make(map[string]*FontBuilder)

	manager.Register("default",
		NewFont(FontFamilySatoshi).
			SetSize(14).
			SetWeight(FontWeightRegular).
			IncludeIcons(true))

	manager.Register("regular",
		NewFont(FontFamilySatoshi).
			SetSize(14).
			SetWeight(FontWeightRegular).
			IncludeIcons(true))

	manager.Register("black",
		NewFont(FontFamilySatoshi).
			SetSize(14).
			SetWeight(FontWeightBlack).
			IncludeIcons(true))

	manager.Register("code",
		NewFont(FontFamilyFiraCode).
			SetSize(14).
			SetWeight(FontWeightRegular).
			IncludeIcons(true))

	manager.Register("bebas",
		NewFont(FontFamilyBebas).
			SetSize(14).
			SetWeight(FontWeightRegular).
			IncludeIcons(false))

	manager.Register("hegarty",
		NewFont(FontFamilyHegarty).
			SetSize(14).
			SetWeight(FontWeightRegular).
			IncludeIcons(false))

	manager.Register("pacifico",
		NewFont(FontFamilyPacifico).
			SetSize(14).
			SetWeight(FontWeightRegular).
			IncludeIcons(false))

	manager.Register("nabla",
		NewFont(FontFamilyNabla).
			SetSize(14).
			SetWeight(FontWeightRegular).
			IncludeIcons(false))

	manager.Register("icons_h4",
		NewFont(FontFamilyLucide).
			SetSize(14).
			SetWeight(FontWeightRegular).
			IncludeIcons(false))

}
