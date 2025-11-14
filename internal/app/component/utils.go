package component

import (
	"image/color"

	"github.com/AllenDang/cimgui-go/imgui"
)

func ToVec4Color(col color.Color) imgui.Vec4 {
	const mask = 0xffff

	r, g, b, a := col.RGBA()

	return imgui.Vec4{
		X: float32(r) / mask,
		Y: float32(g) / mask,
		Z: float32(b) / mask,
		W: float32(a) / mask,
	}
}

// addTextClipped adds text with truncation.
func AddTextClipped(draw *imgui.DrawList, pos imgui.Vec2, maxWidth float32, text string, color imgui.Vec4) {
	if text == "" {
		return
	}
	ts := imgui.CalcTextSizeV(text, false, 0)
	if ts.X <= maxWidth {
		draw.AddTextVec2(pos, imgui.ColorU32Vec4(color), text)
		return
	}
	ellipsis := "..."
	ellipsisW := imgui.CalcTextSizeV(ellipsis, false, 0).X

	left, right := 0, len(text)
	for left < right {
		mid := (left + right) / 2
		sub := text[:mid]
		w := imgui.CalcTextSizeV(sub, false, 0).X + ellipsisW
		if w <= maxWidth {
			left = mid + 1
		} else {
			right = mid
		}
	}
	if right <= 0 {
		draw.AddTextVec2(pos, imgui.ColorU32Vec4(color), ellipsis)
		return
	}
	draw.AddTextVec2(pos, imgui.ColorU32Vec4(color), text[:right-1]+ellipsis)
}
