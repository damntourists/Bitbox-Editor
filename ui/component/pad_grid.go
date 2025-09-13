package component

import (
	"bitbox-editor/lib/preset"
	"bitbox-editor/ui/events"
	"context"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
)

type PadGridComponent struct {
	*Component
	rows, cols int
	pads       [][]*PadComponent
	padSize    int

	preset *preset.Preset
}

func (p *PadGridComponent) SetPreset(preset *preset.Preset) *PadGridComponent {
	p.preset = preset

	return p
}

func (p *PadGridComponent) Layout() {
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: 0})
	defer imgui.PopStyleVarV(2)

	for i := 0; i < p.rows; i++ {
		for j := 0; j < p.cols; j++ {
			p := NewPadComponent(
				imgui.IDStr(fmt.Sprintf("pad-%dx%d", i, j)),
				j,
				i,
				fmt.Sprintf("row: %d", i),
				fmt.Sprintf("col: %d", j),
				"...",
			)
			//p.SetProgress(50.0)
			p.MouseEvents.AddListener(func(ctx context.Context, record events.MouseEventRecord) {
				switch record.Type {
				case events.Clicked:
					log.Info(fmt.Sprintf("mouse event: %v", record.Type))
				}
			}, "pad-mouse-events")

			p.Layout()
			imgui.SameLine()
		}
		imgui.NewLine()
	}

}

func NewPadGrid(id imgui.ID, rows, cols, size int) *PadGridComponent {
	cmp := &PadGridComponent{
		Component: NewComponent(id),
		rows:      rows,
		cols:      cols,
		padSize:   size,
	}
	cmp.Component.layoutBuilder = cmp
	return cmp
}
