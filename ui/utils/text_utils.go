package utils

import "github.com/AllenDang/cimgui-go/imgui"

// AddTextClipped adds text with truncation.
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
