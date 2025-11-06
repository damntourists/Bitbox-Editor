package label

import (
	"bitbox-editor/internal/app/animation" // Keep this import
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/logging"
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

var log = logging.NewLogger("label")

// labelCounter ensures unique IDs for labels with the same text
var labelCounter uint64

// LabelComponent is a simple text label with a background.
type LabelComponent struct {
	*component.Component[*LabelComponent]
}

func NewLabel(text string) *LabelComponent {
	counter := atomic.AddUint64(&labelCounter, 1)
	uniqueID := imgui.IDStr(fmt.Sprintf("%s##label_%d", text, counter))
	return NewLabelWithID(uniqueID, text)
}

func NewLabelWithID(id imgui.ID, text string) *LabelComponent {
	t := theme.GetCurrentTheme()
	baseBgColor := t.Style.Colors.HeaderActive.Vec4
	hoverBgColor := t.Style.Colors.HeaderHovered.Vec4

	cmp := &LabelComponent{}

	// Create the component with the handler
	cmp.Component = component.NewComponent[*LabelComponent](id, cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)

	cmp.SetText(text)
	cmp.SetFgColor(t.Style.Colors.Text.Vec4)
	cmp.SetBgColor(baseBgColor)
	cmp.SetBgHoveredColor(hoverBgColor)
	cmp.SetPadding(2)
	cmp.SetRounding(2)
	cmp.SetOutlineWidth(1)
	cmp.SetOutline(true)
	cmp.SetOutlineColor(t.Style.Colors.Border.Vec4)

	return cmp
}

// handleUpdate processes commands.
func (lc *LabelComponent) handleUpdate(cmd component.UpdateCmd) {
	if lc.Component.HandleGlobalUpdate(cmd) {
		return
	}

	log.Warn(
		"LabelComponent unhandled update",
		zap.String("id", lc.IDStr()),
		zap.Any("cmd", cmd),
	)
}

// Layout renders the label
func (lc *LabelComponent) Layout() {
	lc.Component.ProcessUpdates()

	text := lc.Component.Text()
	padding := lc.Component.Padding()
	rounding := lc.Component.Rounding()
	outline := lc.Component.Outline()
	outlineWidth := lc.Component.OutlineWidth()
	outlineColor := lc.Component.OutlineColor()
	fgColor := lc.Component.TextColor()
	currentBgColor := lc.Component.GetAnimatedBgColor() // Use animated color

	// Get size of text
	displayText := text
	textWithID := text + "##" + strconv.Itoa(int(lc.ID()))
	textSize := imgui.CalcTextSize(displayText)

	// Calculate the size of the entire component
	size := imgui.Vec2{
		X: textSize.X + padding*4,
		Y: textSize.Y + padding*4,
	}

	pos := imgui.CursorScreenPos()

	imgui.InvisibleButtonV(textWithID, size, imgui.ButtonFlagsNone)

	// Calculate bounds for drawing
	maxBounds := imgui.Vec2{
		X: pos.X + size.X,
		Y: pos.Y + size.Y,
	}

	textPos := imgui.Vec2{
		X: pos.X + (size.X-textSize.X)/2,
		Y: pos.Y + (size.Y-textSize.Y)/2,
	}

	dl := imgui.WindowDrawList()

	// Default to pill shape if no rounding is set
	if rounding <= 0 {
		rounding = size.X / 2
	}

	// Draw background
	dl.AddRectFilledV(
		pos,
		maxBounds,
		imgui.ColorU32Vec4(currentBgColor),
		rounding,
		imgui.DrawFlagsNone,
	)

	// Draw outline
	if outline {
		dl.AddRectV(
			pos,
			maxBounds,
			imgui.ColorU32Vec4(outlineColor),
			rounding,
			imgui.DrawFlagsNone,
			outlineWidth,
		)
	}

	// Draw text
	dl.AddTextVec2(
		textPos,
		imgui.ColorU32Vec4(fgColor),
		displayText,
	)
}

func (lc *LabelComponent) SetText(text string) *LabelComponent {
	lc.Component.SetText(text)
	return lc
}

func (lc *LabelComponent) SetBgColor(color imgui.Vec4) *LabelComponent {
	lc.Component.SetBgColor(color)
	return lc
}

func (lc *LabelComponent) SetBgHoveredColor(color imgui.Vec4) *LabelComponent {
	lc.Component.SetBgHoveredColor(color)
	return lc
}

func (lc *LabelComponent) SetFgColor(color imgui.Vec4) *LabelComponent {
	lc.Component.SetFgColor(color)
	return lc
}

func (lc *LabelComponent) SetOutlineColor(color imgui.Vec4) *LabelComponent {
	lc.Component.SetOutlineColor(color)
	return lc
}

func (lc *LabelComponent) SetPadding(padding float32) *LabelComponent {
	lc.Component.SetPadding(padding)
	return lc
}

func (lc *LabelComponent) SetRounding(rounding float32) *LabelComponent {
	lc.Component.SetRounding(rounding)
	return lc
}

func (lc *LabelComponent) SetOutlineWidth(width float32) *LabelComponent {
	lc.Component.SetOutlineWidth(width)
	return lc
}

func (lc *LabelComponent) SetOutline(outline bool) *LabelComponent {
	lc.Component.SetOutline(outline)
	return lc
}

// AnimateToBgColor animates the background color to a target color
func (lc *LabelComponent) AnimateToBgColor(targetColor imgui.Vec4) *LabelComponent {
	lc.Component.StartVec4Animation(
		component.CmdSetBgColor,
		targetColor,
		animation.DefaultColorFadeDuration,
		animation.DefaultEasingFunction,
	)
	return lc
}

// Destroy cleans up the component
func (lc *LabelComponent) Destroy() {
	lc.Component.Destroy()
}
