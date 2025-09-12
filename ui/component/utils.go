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
