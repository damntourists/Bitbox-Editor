package component

import (
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
)

func WavyText(
	text string,
	colormap implot.Colormap,
	outline bool,
	outlineColor uint32,
	waveSpeedMultiplier imgui.Vec2,
	waveMultiplier imgui.Vec2,
	amplitude imgui.Vec2,
	colorSpeedMultiplier float64,
) {
	drawList := imgui.WindowDrawList()
	startPos := imgui.CursorScreenPos()
	time := imgui.Time()

	currentX := startPos.X

	for i, char := range text {
		charStr := string(char)
		charSize := imgui.CalcTextSizeV(charStr, false, 0)

		waveOffsetX := float32(math.Cos(time*float64(waveSpeedMultiplier.X)+float64(i)*0.4)) * amplitude.X
		waveOffsetY := float32(math.Sin(time*float64(waveSpeedMultiplier.Y)+float64(i)*0.4)) * amplitude.Y

		charPos := imgui.Vec2{
			X: currentX + waveOffsetX*waveMultiplier.X + (amplitude.X / 2),
			Y: startPos.Y + waveOffsetY*waveMultiplier.Y + (amplitude.Y / 2),
		}

		if outline {
			outlineThickness := float32(1.0)
			offsets := []imgui.Vec2{
				{X: -outlineThickness, Y: 0},
				{X: outlineThickness, Y: 0},
				{X: 0, Y: -outlineThickness},
				{X: 0, Y: outlineThickness},
				{X: -outlineThickness, Y: -outlineThickness},
				{X: outlineThickness, Y: -outlineThickness},
				{X: -outlineThickness, Y: outlineThickness},
				{X: outlineThickness, Y: outlineThickness},
			}

			for _, offset := range offsets {
				outlinePos := imgui.Vec2{
					X: charPos.X + offset.X,
					Y: charPos.Y + offset.Y,
				}
				drawList.AddTextVec2(outlinePos, outlineColor, charStr)
			}
		}

		var textColor uint32
		t := float32(math.Mod(time*colorSpeedMultiplier+float64(i)*0.02, 1.0))
		color := implot.SampleColormapV(t, colormap)
		textColor = imgui.ColorU32Vec4(color)

		// Draw main character on top
		drawList.AddTextVec2(charPos, textColor, charStr)

		currentX += charSize.X
	}

	totalSize := imgui.CalcTextSizeV(text, false, 0)
	imgui.Dummy(imgui.Vec2{X: totalSize.X, Y: totalSize.Y + 6})
}
