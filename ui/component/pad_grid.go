package component

import (
	"bitbox-editor/lib/preset"
	"bitbox-editor/ui/events"
	"context"
	"fmt"
	"path/filepath"

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

func (p *PadGridComponent) SetRows(rows int) *PadGridComponent {
	p.rows = rows
	return p
}

func (p *PadGridComponent) SetCols(cols int) *PadGridComponent {
	p.cols = cols
	return p
}

func (p *PadGridComponent) SetPadSize(size int) *PadGridComponent {
	p.padSize = size
	return p
}

func (p *PadGridComponent) SetPreset(preset *preset.Preset) *PadGridComponent {
	p.preset = preset
	for _, cell := range preset.BitboxConfig().Session.Cells {
		if cell.Row == nil || cell.Column == nil {
			continue
		}

		pad := p.Pad(*cell.Row, *cell.Column)
		if pad != nil {
			if pad.wave == nil {

				for _, wav := range preset.Wavs() {
					resolvedPath, _ := preset.ResolveFile(cell.Filename)
					if wav.Name == filepath.Base(resolvedPath) {
						pad.SetWave(wav)
					}
				}

			}
		}
	}

	return p
}

func (p *PadGridComponent) Pad(row, col int) *PadComponent {
	for _, pad := range p.pads {
		if pad.Row() == row && pad.Col() == col {
			return pad
		}
	}
	return nil
}

func (p *PadGridComponent) Layout() {
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: 0})
	defer imgui.PopStyleVarV(2)

	for i, pad := range p.pads {
		if p.selectedPad != nil {
			if pad == p.selectedPad {
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

	cmp.Component.layoutBuilder = cmp

	// TODO: Make configurable
	for r := 0; r < cmp.rows; r++ {
		for c := 0; c < cmp.cols; c++ {
			pc := NewPadComponent(
				imgui.IDStr(fmt.Sprintf("pad-%dx%d", r, c)),
				r, c,
				"unset", "", "",
				float32(size),
			)

			pc.MouseEvents().AddListener(func(ctx context.Context, record events.MouseEventRecord) {
				switch record.Type {
				case events.Clicked:
					cmp.selectedPad = pc
					cmp.Events.Emit(ctx, events.PadEventRecord{
						Type: events.PadActivated,
						Data: pc,
					})
				}
			}, fmt.Sprintf("%s-%dx%d-mouse-events", pc.IDStr(), r, c))
			cmp.pads = append(cmp.pads, pc)
		}

	}

	return cmp
}
