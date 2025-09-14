package component

import (
	"bitbox-editor/lib/preset"
	"bitbox-editor/ui/events"
	"context"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

type PadGridComponent struct {
	*Component
	rows, cols int

	pads        []*PadComponent
	padSize     int
	selectedPad *PadComponent

	preset *preset.Preset

	Events signals.Signal[events.PadEventRecord]
}

func (p *PadGridComponent) SetPreset(preset *preset.Preset) *PadGridComponent {
	p.preset = preset
	for _, cell := range preset.BitboxConfig().Session.Cells {

	}

	return p
}

func (p *PadGridComponent) Layout() {
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: 0})
	defer imgui.PopStyleVarV(2)

	for i, pad := range p.pads {
		if p.selectedPad != nil {
			if pad.row == p.selectedPad.row && pad.col == p.selectedPad.col {
				pad.SetSelected(true)
			} else {
				pad.SetSelected(false)
			}
		}
		pad.Layout()

		if i%p.cols != p.cols-1 {
			imgui.SameLine()
		}
	}

}

func NewPadGrid(id imgui.ID, rows, cols, size int) *PadGridComponent {
	cmp := &PadGridComponent{
		Component: NewComponent(id),
		rows:      rows,
		cols:      cols,
		pads:      make([]*PadComponent, 0),
		padSize:   size,
		preset:    nil,
		Events:    signals.New[events.PadEventRecord](),
	}

	for i := 0; i < cmp.rows; i++ {
		for j := 0; j < cmp.cols; j++ {
			pc := NewPadComponent(
				imgui.IDStr(fmt.Sprintf("pad-%dx%d", i, j)),
				i,
				j,
				fmt.Sprintf("row: %d", i),
				fmt.Sprintf("col: %d", j),
				"...",
			)

			pc.MouseEvents.AddListener(func(ctx context.Context, record events.MouseEventRecord) {
				switch record.Type {
				case events.Clicked:
					cmp.selectedPad = pc
					cmp.Events.Emit(ctx, events.PadEventRecord{
						Type: events.PadActivated,
						Data: pc,
					})
				}
			}, "pad-mouse-events")
			cmp.pads = append(cmp.pads, pc)
		}

	}

	cmp.Component.layoutBuilder = cmp
	return cmp
}
