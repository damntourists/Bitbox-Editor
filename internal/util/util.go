package util

import (
	"image"
	"image/color"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
)

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}

// ToVec4Color converts rgba color to imgui.Vec4.
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

// ToVec2 converts image.Point to imgui.Vec2.
func ToVec2(pt image.Point) imgui.Vec2 {
	return imgui.Vec2{
		X: float32(pt.X),
		Y: float32(pt.Y),
	}
}

// Vec4ToRGBA converts imgui's Vec4 to golang rgba color.
func Vec4ToRGBA(vec4 imgui.Vec4) color.RGBA {
	return color.RGBA{
		R: uint8(vec4.X * 255),
		G: uint8(vec4.Y * 255),
		B: uint8(vec4.Z * 255),
		A: uint8(vec4.W * 255),
	}
}
