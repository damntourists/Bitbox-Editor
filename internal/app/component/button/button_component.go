package button

/*
╭──────────────────╮
│ Button Component │
╰──────────────────╯
*/

import (
	"bitbox-editor/internal/app/animation"
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/logging"
	"strconv"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

var log = logging.NewLogger("button")

type Button struct {
	*component.Component[*Button]

	text    string
	enabled bool
	toggled bool

	// Colors for different states
	normalColor   imgui.Vec4
	hoveredColor  imgui.Vec4
	activeColor   imgui.Vec4
	disabledColor imgui.Vec4
	toggledColor  imgui.Vec4

	fgColor      imgui.Vec4
	outlineColor imgui.Vec4

	padding      float32
	rounding     float32
	outline      bool
	outlineWidth float32

	// Fixed size (0 = auto-size based on text)
	fixedWidth  float32
	fixedHeight float32

	// Callbacks
	onClick       func()
	onDoubleClick func()
}

func (b *Button) handleUpdate(cmd component.UpdateCmd) {
	handled := b.Component.HandleGlobalUpdate(cmd)

	// Handle button-specific updates
	switch cmd.Type {
	case component.CmdSetText:
		if text, ok := cmd.Data.(string); ok {
			b.text = text
		}
	case component.CmdSetEnabled:
		if enabled, ok := cmd.Data.(bool); ok {
			b.enabled = enabled
			if enabled {
				b.Component.StartVec4Animation(
					component.CmdSetBgColor,
					b.normalColor,
					animation.DefaultColorFadeDuration,
				)
			} else {
				b.Component.StartVec4Animation(
					component.CmdSetBgColor,
					b.disabledColor,
					animation.DefaultColorFadeDuration,
				)
			}
		}

	default:
		if !handled {
			log.Warn("Button unhandled update", zap.Any("cmd", cmd))
		}
	}
}

func (b *Button) Layout() {
	// Get the current animated background color
	currentBgColor := b.Component.GetAnimatedBgColor()

	// Get size of text
	textSize := imgui.CalcTextSize(b.text)

	// Calculate size of the entire component including all padding
	size := imgui.Vec2{
		X: textSize.X + b.padding*4,
		Y: textSize.Y + b.padding*4,
	}

	// Apply fixed dimensions if set
	if b.fixedWidth > 0 {
		size.X = b.fixedWidth
	}
	if b.fixedHeight > 0 {
		size.Y = b.fixedHeight
	}

	// Get the cursor position where the button will be placed
	pos := imgui.CursorScreenPos()

	// Create invisible button for hover detection
	buttonID := strconv.Itoa(int(b.ID()))
	clicked := imgui.InvisibleButtonV(buttonID, size, imgui.ButtonFlagsNone)

	// Handle click if enabled
	if clicked && b.enabled && b.onClick != nil {
		b.onClick()
	}

	// Check hover state
	isHovered := imgui.IsItemHovered()
	isActive := imgui.IsItemActive()

	//Determine target color based on state
	var targetColor imgui.Vec4
	if !b.enabled {
		targetColor = b.disabledColor
	} else if b.toggled {
		// When toggled, use toggled color (or active color with slight variation)
		if isActive {
			targetColor = b.activeColor
		} else if isHovered {
			// Slightly brighten the toggled color on hover
			targetColor = imgui.Vec4{
				X: b.toggledColor.X * 1.1,
				Y: b.toggledColor.Y * 1.1,
				Z: b.toggledColor.Z * 1.1,
				W: b.toggledColor.W,
			}
		} else {
			targetColor = b.toggledColor
		}
	} else if isActive {
		targetColor = b.activeColor
	} else if isHovered {
		targetColor = b.hoveredColor
	} else {
		targetColor = b.normalColor
	}

	b.Component.StartVec4Animation(
		component.CmdSetBgColor,
		targetColor,
		animation.DefaultColorFadeDuration,
	)

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

	// Draw background
	dl.AddRectFilledV(
		pos,
		maxBounds,
		imgui.ColorU32Vec4(currentBgColor),
		b.rounding,
		imgui.DrawFlagsNone,
	)

	// Draw outline
	if b.outline {
		dl.AddRectV(
			pos,
			maxBounds,
			imgui.ColorU32Vec4(b.outlineColor),
			b.rounding,
			imgui.DrawFlagsNone,
			b.outlineWidth,
		)
	}

	// Draw text (with dimmed color if disabled)
	textColor := b.fgColor
	if !b.enabled {
		textColor = imgui.Vec4{
			X: b.fgColor.X * 0.5,
			Y: b.fgColor.Y * 0.5,
			Z: b.fgColor.Z * 0.5,
			W: b.fgColor.W,
		}
	}

	dl.AddTextVec2(
		textPos,
		imgui.ColorU32Vec4(textColor),
		b.text,
	)

	// Set cursor for enabled buttons
	if b.enabled && (isHovered || isActive) {
		imgui.SetMouseCursor(imgui.MouseCursorHand)
	}
}

func (b *Button) SetText(text string) *Button {
	b.text = text
	return b
}

func (b *Button) SetEnabled(enabled bool) *Button {
	b.enabled = enabled
	return b
}

func (b *Button) SetNormalColor(color imgui.Vec4) *Button {
	b.normalColor = color
	b.Component.SetBgColor(color)
	return b
}

func (b *Button) SetHoveredColor(color imgui.Vec4) *Button {
	b.hoveredColor = color
	b.Component.SetBgHoveredColor(color)
	return b
}

func (b *Button) SetActiveColor(color imgui.Vec4) *Button {
	b.activeColor = color
	return b
}

func (b *Button) SetDisabledColor(color imgui.Vec4) *Button {
	b.disabledColor = color
	return b
}

func (b *Button) SetFgColor(color imgui.Vec4) *Button {
	b.fgColor = color
	b.Component.SetFgColor(color)
	return b
}

func (b *Button) SetOutlineColor(color imgui.Vec4) *Button {
	b.outlineColor = color
	b.Component.SetOutlineColor(color)
	return b
}

func (b *Button) SetPadding(padding float32) *Button {
	b.padding = padding
	b.Component.SetPadding(padding)
	return b
}

func (b *Button) SetRounding(rounding float32) *Button {
	b.rounding = rounding
	b.Component.SetRounding(rounding)
	return b
}

func (b *Button) SetOutlineWidth(width float32) *Button {
	b.outlineWidth = width
	b.Component.SetOutlineWidth(width)
	return b
}

func (b *Button) SetOutline(outline bool) *Button {
	b.outline = outline
	b.Component.SetOutline(outline)
	return b
}

func (b *Button) SetOnClick(callback func()) *Button {
	b.onClick = callback
	return b
}

func (b *Button) SetOnDoubleClick(callback func()) *Button {
	b.onDoubleClick = callback
	return b
}

func (b *Button) SetFixedWidth(width float32) *Button {
	b.fixedWidth = width
	return b
}

func (b *Button) SetFixedHeight(height float32) *Button {
	b.fixedHeight = height
	return b
}

func (b *Button) SetFixedSize(width, height float32) *Button {
	b.fixedWidth = width
	b.fixedHeight = height
	return b
}

func (b *Button) SetToggled(toggled bool) *Button {
	b.toggled = toggled
	return b
}

func (b *Button) IsToggled() bool {
	return b.toggled
}

func (b *Button) SetToggledColor(color imgui.Vec4) *Button {
	b.toggledColor = color
	return b
}

func NewButton(text string) *Button {
	return NewButtonWithID(imgui.IDStr(text), text)
}

func NewButtonWithID(id imgui.ID, text string) *Button {
	t := theme.GetCurrentTheme()

	// Use theme colors for different states
	normalColor := t.Style.Colors.Button.Vec4
	hoveredColor := t.Style.Colors.ButtonHovered.Vec4
	activeColor := t.Style.Colors.ButtonActive.Vec4
	toggledColor := t.Style.Colors.TabHovered.Vec4
	disabledColor := imgui.Vec4{
		X: normalColor.X * 0.6,
		Y: normalColor.Y * 0.6,
		Z: normalColor.Z * 0.6,
		W: normalColor.W,
	}

	cmp := &Button{
		text:    text,
		enabled: true,
		toggled: false,

		normalColor:   normalColor,
		hoveredColor:  hoveredColor,
		activeColor:   activeColor,
		disabledColor: disabledColor,
		toggledColor:  toggledColor,

		fgColor: t.Style.Colors.Text.Vec4,

		padding:      8,
		rounding:     4,
		outlineWidth: 1,
		outline:      true,
		outlineColor: t.Style.Colors.Border.Vec4,
	}

	cmp.Component = component.NewComponent[*Button](id, cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)

	cmp.Component.SetBgColor(normalColor)
	cmp.Component.SetBgHoveredColor(hoveredColor)

	return cmp
}
