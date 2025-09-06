package util

import "github.com/AllenDang/cimgui-go/imgui"

func Vec2Sort(a, b imgui.Vec2) int {
	if (a.X * a.Y) < (b.X * b.Y) {
		return 1
	}
	return 0
}
