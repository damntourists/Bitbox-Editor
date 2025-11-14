package theme

import (
	"bitbox-editor/internal/util"
	"errors"
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
)

type (
	ThemeCollection struct {
		Themes []*Theme
	}

	Theme struct {
		Name        string    `toml:"name"`
		Author      string    `toml:"author"`
		Description string    `toml:"description"`
		Tags        []string  `toml:"tags"`
		Date        time.Time `toml:"date"`
		Style       Style
	}

	Style struct {
		Alpha                     float32   `toml:"alpha"`
		DisabledAlpha             float32   `toml:"disabledAlpha"`
		WindowPadding             []float32 `toml:"windowPadding"`
		WindowRounding            float32   `toml:"windowRounding"`
		WindowBorderSize          float32   `toml:"windowBorderSize"`
		WindowMinSize             []float32 `toml:"windowMinSize"`
		WindowTitleAlign          []float32 `toml:"windowTitleAlign"`
		WindowMenuButtonPosition  string    `toml:"windowMenuButtonPosition"`
		ChildRounding             float32   `toml:"childRounding"`
		ChildBorderSize           float32   `toml:"childBorderSize"`
		PopupRounding             float32   `toml:"popupRounding"`
		PopupBorderSize           float32   `toml:"popupBorderSize"`
		FramePadding              []float32 `toml:"framePadding"`
		FrameRounding             float32   `toml:"frameRounding"`
		FrameBorderSize           float32   `toml:"frameBorderSize"`
		ItemSpacing               []float32 `toml:"itemSpacing"`
		ItemInnerSpacing          []float32 `toml:"itemInnerSpacing"`
		CellPadding               []float32 `toml:"cellPadding"`
		IndentSpacing             float32   `toml:"indentSpacing"`
		ColumnsMinSpacing         float32   `toml:"columnsMinSpacing"`
		ScrollbarSize             float32   `toml:"scrollbarSize"`
		ScrollbarRounding         float32   `toml:"scrollbarRounding"`
		GrabMinSize               float32   `toml:"grabMinSize"`
		GrabRounding              float32   `toml:"grabRounding"`
		TabRounding               float32   `toml:"tabRounding"`
		TabBorderSize             float32   `toml:"tabBorderSize"`
		TabMinWidthForCloseButton float32   `toml:"tabMinWidthForCloseButton"`
		ColorButtonPosition       string    `toml:"colorButtonPosition"`
		ButtonTextAlign           []float32 `toml:"buttonTextAlign"`
		SelectableTextAlign       []float32 `toml:"selectableTextAlign"`

		Colors StyleColor
	}

	StyleColor struct {
		Text                  rgba
		TextDisabled          rgba
		WindowBg              rgba
		ChildBg               rgba
		PopupBg               rgba
		Border                rgba
		BorderShadow          rgba
		FrameBg               rgba
		FrameBgHovered        rgba
		FrameBgActive         rgba
		TitleBg               rgba
		TitleBgActive         rgba
		TitleBgCollapsed      rgba
		MenuBarBg             rgba
		ScrollbarBg           rgba
		ScrollbarGrab         rgba
		ScrollbarGrabHovered  rgba
		ScrollbarGrabActive   rgba
		CheckMark             rgba
		SliderGrab            rgba
		SliderGrabActive      rgba
		Button                rgba
		ButtonHovered         rgba
		ButtonActive          rgba
		Header                rgba
		HeaderHovered         rgba
		HeaderActive          rgba
		Separator             rgba
		SeparatorHovered      rgba
		SeparatorActive       rgba
		ResizeGrip            rgba
		ResizeGripHovered     rgba
		ResizeGripActive      rgba
		Tab                   rgba
		TabHovered            rgba
		TabActive             rgba
		TabUnfocused          rgba
		TabUnfocusedActive    rgba
		DockingPreview        rgba
		DockingEmptyBg        rgba
		PlotLines             rgba
		PlotLinesHovered      rgba
		PlotHistogram         rgba
		PlotHistogramHovered  rgba
		TableHeaderBg         rgba
		TableBorderStrong     rgba
		TableBorderLight      rgba
		TableRowBg            rgba
		TableRowBgAlt         rgba
		TextSelectedBg        rgba
		DragDropTarget        rgba
		NavHighlight          rgba
		NavWindowingHighlight rgba
		NavWindowingDimBg     rgba
		ModalWindowDimBg      rgba
	}

	rgba struct {
		imgui.Vec4
	}
)

func (c *rgba) AsRGBA() color.RGBA {
	// Vec4ToRGBA converts imgui's Vec4 to golang rgba color.
	return color.RGBA{
		R: uint8(c.Vec4.X * 255),
		G: uint8(c.Vec4.Y * 255),
		B: uint8(c.Vec4.Z * 255),
		A: uint8(c.Vec4.W * 255),
	}

}

func (c *rgba) AsUint32() uint32 {
	col := c.AsRGBA()
	mask := uint32(0xff)

	return uint32(col.R)&mask +
		uint32(col.G)&mask<<8 +
		uint32(col.B)&mask<<16 +
		uint32(col.A)&mask<<24
}

func (c *rgba) UnmarshalText(text []byte) error {
	var err error
	r := regexp.MustCompile(
		`.*rgba\((\d+|\d+\.\d+),\s*(\d+|\d+\.\d+),\s*(\d+|\d+\.\d+),\s*(\d+|\d+\.\d+)*\)`,
	)

	res := r.FindAllStringSubmatch(string(text), -1)
	pR, _ := strconv.ParseFloat(res[0][1], 64)
	pG, _ := strconv.ParseFloat(res[0][2], 64)
	pB, _ := strconv.ParseFloat(res[0][3], 64)
	pA, _ := strconv.ParseFloat(res[0][4], 64)

	// check if we need to divide
	if pR <= 1 && pG <= 1 && pB <= 1 && pA <= 1 {
		c.Vec4 = imgui.Vec4{
			X: float32(pR),
			Y: float32(pG),
			Z: float32(pB),
			W: float32(pA),
		}
	} else {
		c.Vec4 = imgui.Vec4{
			X: float32(pR) / 255,
			Y: float32(pG) / 255,
			Z: float32(pB) / 255,
			W: float32(pA),
		}
	}

	return err
}

func SetCurrentTheme(theme *Theme) {
	currentTheme = theme
}

func GetCurrentTheme() *Theme {
	if currentTheme == nil {
		var err error
		currentTheme, err = GetThemeByName(DEFAULT_THEME)
		util.PanicOnError(err)
	}

	return currentTheme
}

func SetCurrentColormap(cmap implot.Colormap) {
	currentColorMap = cmap
}

func GetCurrentColormap() implot.Colormap {
	return currentColorMap
}

// GetColormapByName returns the colormap enum value for the given name
func GetColormapByName(name string) implot.Colormap {
	switch strings.ToLower(name) {
	case "deep":
		return implot.ColormapDeep
	case "dark":
		return implot.ColormapDark
	case "pastel":
		return implot.ColormapPastel
	case "paired":
		return implot.ColormapPaired
	case "viridis":
		return implot.ColormapViridis
	case "plasma":
		return implot.ColormapPlasma
	case "hot":
		return implot.ColormapHot
	case "cool":
		return implot.ColormapCool
	case "pink":
		return implot.ColormapPink
	case "jet":
		return implot.ColormapJet
	case "twilight":
		return implot.ColormapTwilight
	case "rdbu":
		return implot.ColormapRdBu
	case "brbg":
		return implot.ColormapBrBG
	case "piyg":
		return implot.ColormapPiYG
	case "spectral":
		return implot.ColormapSpectral
	case "greys":
		return implot.ColormapGreys
	default:
		return implot.ColormapCool // Default fallback
	}
}

// GetColormapName returns the string name for the given colormap enum
func GetColormapName(cmap implot.Colormap) string {
	switch cmap {
	case implot.ColormapDeep:
		return "Deep"
	case implot.ColormapDark:
		return "Dark"
	case implot.ColormapPastel:
		return "Pastel"
	case implot.ColormapPaired:
		return "Paired"
	case implot.ColormapViridis:
		return "Viridis"
	case implot.ColormapPlasma:
		return "Plasma"
	case implot.ColormapHot:
		return "Hot"
	case implot.ColormapCool:
		return "Cool"
	case implot.ColormapPink:
		return "Pink"
	case implot.ColormapJet:
		return "Jet"
	case implot.ColormapTwilight:
		return "Twilight"
	case implot.ColormapRdBu:
		return "RdBu"
	case implot.ColormapBrBG:
		return "BrBG"
	case implot.ColormapPiYG:
		return "PiYG"
	case implot.ColormapSpectral:
		return "Spectral"
	case implot.ColormapGreys:
		return "Greys"
	default:
		return "Cool"
	}
}

// GetAllColormapNames returns a list of all available colormap names
func GetAllColormapNames() []string {
	return []string{
		"Deep", "Dark", "Pastel", "Paired", "Viridis", "Plasma",
		"Hot", "Cool", "Pink", "Jet", "Twilight", "RdBu",
		"BrBG", "PiYG", "Spectral", "Greys",
	}
}

func (t *Theme) Apply() func() {
	style := imgui.CurrentStyle()
	style.SetAlpha(t.Style.Alpha)
	style.SetDisabledAlpha(t.Style.DisabledAlpha)
	style.SetWindowPadding(convertFloatSlice(t.Style.WindowPadding))
	style.SetWindowRounding(t.Style.WindowRounding)
	style.SetWindowBorderSize(t.Style.WindowBorderSize)
	style.SetWindowMinSize(convertFloatSlice(t.Style.WindowMinSize))
	style.SetWindowTitleAlign(convertFloatSlice(t.Style.WindowTitleAlign))
	style.SetWindowMenuButtonPosition(convertStringToDir(t.Style.WindowMenuButtonPosition))
	style.SetChildRounding(t.Style.ChildRounding)
	style.SetChildBorderSize(t.Style.ChildBorderSize)
	style.SetPopupRounding(t.Style.PopupRounding)
	style.SetPopupBorderSize(t.Style.PopupBorderSize)
	style.SetFramePadding(convertFloatSlice(t.Style.FramePadding))
	style.SetFrameRounding(t.Style.FrameRounding)
	style.SetFrameBorderSize(t.Style.FrameBorderSize)
	style.SetItemSpacing(convertFloatSlice(t.Style.ItemSpacing))
	style.SetItemInnerSpacing(convertFloatSlice(t.Style.ItemInnerSpacing))
	style.SetCellPadding(convertFloatSlice(t.Style.CellPadding))
	style.SetIndentSpacing(t.Style.IndentSpacing)
	style.SetColumnsMinSpacing(t.Style.ColumnsMinSpacing)
	style.SetScrollbarSize(t.Style.ScrollbarSize)
	style.SetScrollbarRounding(t.Style.ScrollbarRounding)
	style.SetGrabMinSize(t.Style.GrabMinSize)
	style.SetGrabRounding(t.Style.GrabRounding)
	style.SetTabRounding(t.Style.TabRounding)
	style.SetTabBorderSize(t.Style.TabBorderSize)
	style.SetTabCloseButtonMinWidthSelected(t.Style.TabMinWidthForCloseButton)
	style.SetTabCloseButtonMinWidthUnselected(t.Style.TabMinWidthForCloseButton)
	style.SetColorButtonPosition(convertStringToDirColorButton(t.Style.ColorButtonPosition))
	style.SetButtonTextAlign(convertFloatSlice(t.Style.ButtonTextAlign))
	style.SetSelectableTextAlign(convertFloatSlice(t.Style.SelectableTextAlign))

	// colors

	imgui.PushStyleColorVec4(imgui.ColText, t.Style.Colors.Text.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTextDisabled, t.Style.Colors.TextDisabled.Vec4)
	imgui.PushStyleColorVec4(imgui.ColWindowBg, t.Style.Colors.WindowBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColChildBg, t.Style.Colors.ChildBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColPopupBg, t.Style.Colors.PopupBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColBorder, t.Style.Colors.Border.Vec4)
	imgui.PushStyleColorVec4(imgui.ColBorderShadow, t.Style.Colors.BorderShadow.Vec4)
	imgui.PushStyleColorVec4(imgui.ColFrameBg, t.Style.Colors.FrameBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, t.Style.Colors.FrameBgHovered.Vec4)
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, t.Style.Colors.FrameBgActive.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTitleBg, t.Style.Colors.TitleBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTitleBgActive, t.Style.Colors.TitleBgActive.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTitleBgCollapsed, t.Style.Colors.TitleBgCollapsed.Vec4)
	imgui.PushStyleColorVec4(imgui.ColMenuBarBg, t.Style.Colors.MenuBarBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColScrollbarBg, t.Style.Colors.ScrollbarBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColScrollbarGrab, t.Style.Colors.ScrollbarGrab.Vec4)
	imgui.PushStyleColorVec4(imgui.ColScrollbarGrabHovered, t.Style.Colors.ScrollbarGrabHovered.Vec4)
	imgui.PushStyleColorVec4(imgui.ColScrollbarGrabActive, t.Style.Colors.ScrollbarGrabActive.Vec4)
	imgui.PushStyleColorVec4(imgui.ColCheckMark, t.Style.Colors.CheckMark.Vec4)
	imgui.PushStyleColorVec4(imgui.ColSliderGrab, t.Style.Colors.SliderGrab.Vec4)
	imgui.PushStyleColorVec4(imgui.ColSliderGrabActive, t.Style.Colors.SliderGrabActive.Vec4)
	imgui.PushStyleColorVec4(imgui.ColButton, t.Style.Colors.Button.Vec4)
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, t.Style.Colors.ButtonHovered.Vec4)
	imgui.PushStyleColorVec4(imgui.ColButtonActive, t.Style.Colors.ButtonActive.Vec4)
	imgui.PushStyleColorVec4(imgui.ColHeader, t.Style.Colors.Header.Vec4)
	imgui.PushStyleColorVec4(imgui.ColHeaderHovered, t.Style.Colors.HeaderHovered.Vec4)
	imgui.PushStyleColorVec4(imgui.ColHeaderActive, t.Style.Colors.HeaderActive.Vec4)
	imgui.PushStyleColorVec4(imgui.ColSeparator, t.Style.Colors.Separator.Vec4)
	imgui.PushStyleColorVec4(imgui.ColSeparatorHovered, t.Style.Colors.SeparatorHovered.Vec4)
	imgui.PushStyleColorVec4(imgui.ColSeparatorActive, t.Style.Colors.SeparatorActive.Vec4)
	imgui.PushStyleColorVec4(imgui.ColResizeGrip, t.Style.Colors.ResizeGrip.Vec4)
	imgui.PushStyleColorVec4(imgui.ColResizeGripHovered, t.Style.Colors.ResizeGripHovered.Vec4)
	imgui.PushStyleColorVec4(imgui.ColResizeGripActive, t.Style.Colors.ResizeGripActive.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTab, t.Style.Colors.Tab.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTabHovered, t.Style.Colors.TabHovered.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTabSelected, t.Style.Colors.TabActive.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTabDimmed, t.Style.Colors.TabUnfocused.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTabDimmedSelected, t.Style.Colors.TabUnfocusedActive.Vec4)

	imgui.PushStyleColorVec4(imgui.ColDockingEmptyBg, t.Style.Colors.DockingEmptyBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColDockingPreview, t.Style.Colors.DockingPreview.Vec4)

	imgui.PushStyleColorVec4(imgui.ColPlotLines, t.Style.Colors.PlotLines.Vec4)
	imgui.PushStyleColorVec4(imgui.ColPlotLinesHovered, t.Style.Colors.PlotLinesHovered.Vec4)
	imgui.PushStyleColorVec4(imgui.ColPlotHistogram, t.Style.Colors.PlotHistogram.Vec4)
	imgui.PushStyleColorVec4(imgui.ColPlotHistogramHovered, t.Style.Colors.PlotHistogramHovered.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTableHeaderBg, t.Style.Colors.TableHeaderBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTableBorderStrong, t.Style.Colors.TableBorderStrong.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTableBorderLight, t.Style.Colors.TableBorderLight.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTableRowBg, t.Style.Colors.TableRowBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTableRowBgAlt, t.Style.Colors.TableRowBgAlt.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTextSelectedBg, t.Style.Colors.TextSelectedBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColDragDropTarget, t.Style.Colors.DragDropTarget.Vec4)
	imgui.PushStyleColorVec4(imgui.ColNavCursor, t.Style.Colors.NavHighlight.Vec4)
	imgui.PushStyleColorVec4(imgui.ColNavWindowingHighlight, t.Style.Colors.NavWindowingHighlight.Vec4)
	imgui.PushStyleColorVec4(imgui.ColNavWindowingDimBg, t.Style.Colors.NavWindowingDimBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColModalWindowDimBg, t.Style.Colors.ModalWindowDimBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTextLink, t.Style.Colors.NavHighlight.Vec4)
	imgui.PushStyleColorVec4(imgui.ColDockingPreview, t.Style.Colors.DragDropTarget.Vec4)
	imgui.PushStyleColorVec4(imgui.ColDockingEmptyBg, t.Style.Colors.ModalWindowDimBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTabDimmedSelectedOverline, t.Style.Colors.Button.Vec4)
	imgui.PushStyleColorVec4(imgui.ColTabSelectedOverline, t.Style.Colors.ButtonActive.Vec4)
	//ColCOUNT            Col = 58

	//implot.PushStyleColorVec4(implot.ColLine, t.Style.Colors.PlotLines.Vec4)
	//implot.PushStyleColorVec4(implot.ColSelection, t.Style.Colors.PlotLines.Vec4)
	implot.PushStyleVarFloat(implot.StyleVarFillAlpha, 0.7)

	/*
		TODO: Implement implot styles that we can set from main theme

		PlotColLine          PlotCol = 0
		PlotColFill          PlotCol = 1
		PlotColMarkerOutline PlotCol = 2
		PlotColMarkerFill    PlotCol = 3
		PlotColErrorBar      PlotCol = 4
		PlotColFrameBg       PlotCol = 5
		PlotColPlotBg        PlotCol = 6
		PlotColPlotBorder    PlotCol = 7
		PlotColLegendBg      PlotCol = 8
		PlotColLegendBorder  PlotCol = 9
		PlotColLegendText    PlotCol = 10
		PlotColTitleText     PlotCol = 11
		PlotColInlayText     PlotCol = 12
		PlotColAxisText      PlotCol = 13
		PlotColAxisGrid      PlotCol = 14
		PlotColAxisTick      PlotCol = 15
		PlotColAxisBg        PlotCol = 16
		PlotColAxisBgHovered PlotCol = 17
		PlotColAxisBgActive  PlotCol = 18
		PlotColSelection     PlotCol = 19
		PlotColCrosshairs    PlotCol = 20
		PlotColCOUNT         PlotCol = 21

		PlotStyleVarLineWeight         PlotStyleVar = 0
		PlotStyleVarMarker             PlotStyleVar = 1
		PlotStyleVarMarkerSize         PlotStyleVar = 2
		PlotStyleVarMarkerWeight       PlotStyleVar = 3
		PlotStyleVarFillAlpha          PlotStyleVar = 4
		PlotStyleVarErrorBarSize       PlotStyleVar = 5
		PlotStyleVarErrorBarWeight     PlotStyleVar = 6
		PlotStyleVarDigitalBitHeight   PlotStyleVar = 7
		PlotStyleVarDigitalBitGap      PlotStyleVar = 8
		PlotStyleVarPlotBorderSize     PlotStyleVar = 9
		PlotStyleVarMinorAlpha         PlotStyleVar = 10
		PlotStyleVarMajorTickLen       PlotStyleVar = 11
		PlotStyleVarMinorTickLen       PlotStyleVar = 12
		PlotStyleVarMajorTickSize      PlotStyleVar = 13
		PlotStyleVarMinorTickSize      PlotStyleVar = 14
		PlotStyleVarMajorGridSize      PlotStyleVar = 15
		PlotStyleVarMinorGridSize      PlotStyleVar = 16
		PlotStyleVarPlotPadding        PlotStyleVar = 17
		PlotStyleVarLabelPadding       PlotStyleVar = 18
		PlotStyleVarLegendPadding      PlotStyleVar = 19
		PlotStyleVarLegendInnerPadding PlotStyleVar = 20
		PlotStyleVarLegendSpacing      PlotStyleVar = 21
		PlotStyleVarMousePosPadding    PlotStyleVar = 22
		PlotStyleVarAnnotationPadding  PlotStyleVar = 23
		PlotStyleVarFitPadding         PlotStyleVar = 24
		PlotStyleVarPlotDefaultSize    PlotStyleVar = 25
		PlotStyleVarPlotMinSize        PlotStyleVar = 26
		PlotStyleVarCOUNT              PlotStyleVar = 27
	*/

	return func() {
		imgui.PopStyleColorV(int32(imgui.ColCOUNT))

		implot.PopStyleVarV(1)
		//implot.PopStyleColorV(2)
	}
}

func GetNames() []string {
	names := make([]string, 0)
	for _, theme := range Themes.Themes {
		names = append(names, theme.Name)
	}
	return names
}

func GetThemeByName(name string) (*Theme, error) {
	var err error
	var t *Theme
	for _, theme := range Themes.Themes {
		if strings.ToLower(name) == strings.ToLower(theme.Name) {
			t = theme
			return t, nil
		}
	}

	err = errors.New(fmt.Sprintf("Failed to find theme '%s'.", name))
	return nil, err
}

func convertFloatSlice(floats []float32) imgui.Vec2 {
	return imgui.Vec2{X: floats[0], Y: floats[1]}
}

func convertStringToDir(val string) imgui.Dir {
	switch strings.ToLower(val) {
	case "up":
		return imgui.DirUp
	case "down":
		return imgui.DirDown
	case "left":
		return imgui.DirLeft
	case "right":
		return imgui.DirRight
	case "none":
		return imgui.DirNone
	case "":
		return imgui.DirRight // Default
	default:
		panic("Could not convert:" + val)
	}
}

func convertStringToDirColorButton(val string) imgui.Dir {
	switch strings.ToLower(val) {
	case "left":
		return imgui.DirLeft
	case "right":
		return imgui.DirRight
	case "none", "": // ColorButtonPosition doesn't support None, default to Right
		return imgui.DirRight
	default:
		panic("Could not convert ColorButtonPosition:" + val)
	}
}
