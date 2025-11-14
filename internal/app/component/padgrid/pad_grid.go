package padgrid

/*
╔════════════════════╗
║ Pad Grid Component ║
╚════════════════════╝

Layouts

	4x2
	╭───╮╭───╮╭───╮╭───╮
	│0,0││0,1││0,2││0,3│
	╰───╯╰───╯╰───╯╰───╯
	╭───╮╭───╮╭───╮╭───╮
	│1,0││1,1││1,2││1,3│
	╰───╯╰───╯╰───╯╰───╯

	4x4
	╭───╮╭───╮╭───╮╭───╮
	│0,0││0,1││0,2││0,3│
	╰───╯╰───╯╰───╯╰───╯
	╭───╮╭───╮╭───╮╭───╮
	│1,0││1,1││1,2││1,3│
	╰───╯╰───╯╰───╯╰───╯
	╭───╮╭───╮╭───╮╭───╮
	│2,0││2,1││2,2││2,3│
	╰───╯╰───╯╰───╯╰───╯
	╭───╮╭───╮╭───╮╭───╮
	│3,0││3,1││3,2││3,3│
	╰───╯╰───╯╰───╯╰───╯

*/
import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/component/pad"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/audio"
	"bitbox-editor/internal/logging"
	"bitbox-editor/internal/preset"
	"fmt"
	"path/filepath"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

// TODO: Finish documentation

var log = logging.NewLogger("padgrid")

type PadGridConfigPayload struct {
	Rows    int
	Cols    int
	PadSize int
}

type PadGridComponent struct {
	*component.Component[*PadGridComponent]
	rows, cols, padSize int

	pads                []*pad.PadComponent
	selectedPad         *pad.PadComponent
	lastPublishedPadPtr *pad.PadComponent

	preset *preset.Preset

	eventSub chan events.Event
	ownerID  string
}

func NewPadGrid(id imgui.ID, rows, cols, size int) *PadGridComponent {
	cmp := &PadGridComponent{
		rows:     rows,
		cols:     cols,
		padSize:  size,
		preset:   nil,
		eventSub: make(chan events.Event, 50),
	}

	cmp.Component = component.NewComponent[*PadGridComponent](id, cmp.handleUpdate)

	cmp.initPads()

	cmp.Component.SetLayoutBuilder(cmp)

	eventbus.Bus.Subscribe(events.ComponentClickEventKey, cmp.UUID(), cmp.eventSub)

	return cmp
}

// drainEvents translates global bus events into local commands
func (c *PadGridComponent) drainEvents() {
	for {
		select {
		case event := <-c.eventSub:
			var cmd component.UpdateCmd
			switch event.Type() {
			case events.ComponentClickEventKey:
				cmd = component.UpdateCmd{Type: cmdHandlePadClick, Data: event}
			}

			if cmd.Type != 0 {
				c.SendUpdate(cmd)
			}
		default:
			return
		}
	}
}

func (c *PadGridComponent) handleUpdate(cmd component.UpdateCmd) {
	switch ct := cmd.Type.(type) {
	case component.GlobalCommand:
		c.Component.HandleGlobalUpdate(cmd)
		return

	case localCommand:
		switch ct {
		case cmdSetPadGridConfig:
			if payload, ok := cmd.Data.(PadGridConfigPayload); ok {
				needsReinit := c.rows != payload.Rows ||
					c.cols != payload.Cols ||
					c.padSize != payload.PadSize
				c.rows = payload.Rows
				c.cols = payload.Cols
				c.padSize = payload.PadSize
				if needsReinit {
					c.initPads()
					if c.preset != nil {
						c.applyPresetData()
					}
				}
			}

		case cmdSetPadGridPreset:
			if newPreset, ok := cmd.Data.(*preset.Preset); ok {
				c.preset = newPreset
				c.applyPresetData()
			}

		case cmdHandlePadClick:
			// Listens for clicks from the global bus
			if event, ok := cmd.Data.(events.MouseEventRecord); ok {
				if clickedPad, ok := event.Data.(*pad.PadComponent); ok {
					isMyPad := false
					for _, p := range c.pads {
						if p == clickedPad {
							isMyPad = true
							break
						}
					}

					if isMyPad {
						// Clear the last published pad pointer to allow re-triggering
						c.lastPublishedPadPtr = nil
						c.selectedPad = clickedPad
					}
				}
			}
		default:
			log.Warn(
				"PadGridComponent unhandled local command",
				zap.String("id", c.IDStr()),
				zap.Any("cmd", cmd),
			)
		}
		return

	default:
		log.Warn(
			"PadGridComponent unhandled update type",
			zap.String("id", c.IDStr()),
			zap.Any("type", fmt.Sprintf("%T", cmd.Type)),
		)
	}
}

func (c *PadGridComponent) initPads() {
	for _, p := range c.pads {
		p.Destroy()
	}

	c.pads = make([]*pad.PadComponent, 0, c.rows*c.cols)
	c.selectedPad = nil

	for row := 0; row < c.rows; row++ {
		for col := 0; col < c.cols; col++ {
			pc := pad.NewPad(
				imgui.IDStr(fmt.Sprintf("pad-%s-%dx%d", c.UUID(), row, col)),
				row, col,
				float32(c.padSize),
			)
			c.pads = append(c.pads, pc)
		}
	}
}

func (c *PadGridComponent) applyPresetData() {
	if c.preset == nil {
		return
	}

	config := c.preset.BitboxConfig()
	wavMap := make(map[string]*audio.WaveFile)
	for _, wav := range c.preset.Wavs() {
		wavMap[wav.Name] = wav
	}

	for _, cell := range config.Session.Cells {
		if cell.Row == nil || cell.Column == nil {
			continue
		}

		p := c.Pad(*cell.Row, *cell.Column)
		if p == nil {
			continue
		}

		resolvedPath, err := c.preset.ResolveFile(cell.Filename)
		if err != nil {
			log.Warn("Could not resolve file for cell", zap.Error(err))
			continue
		}
		wavName := filepath.Base(resolvedPath)
		if wav, found := wavMap[wavName]; found {
			c.loadAndAssignWaveData(p, wav)
		}
	}
}

func (c *PadGridComponent) loadAndAssignWaveData(pad *pad.PadComponent, wav *audio.WaveFile) {
	dataToSend := audio.WaveDisplayData{
		Name:       wav.Name,
		Path:       wav.Path,
		IsReady:    true,
		LoadFailed: false,
		IsLoading:  false,
		IsPlaying:  false,
		SampleRate: wav.SampleRate,
		NumSamples: 0,
	}
	pad.SetWaveDisplayData(dataToSend)
}

func (c *PadGridComponent) GetPadSize() int {
	return c.padSize
}

func (c *PadGridComponent) Pad(row, col int) *pad.PadComponent {
	for _, p := range c.pads {
		if p.Row() == row && p.Col() == col {
			return p
		}
	}
	return nil
}

func (c *PadGridComponent) Pads() []*pad.PadComponent {
	return c.pads
}

func (c *PadGridComponent) SetRows(rows int) *PadGridComponent {
	payload := PadGridConfigPayload{Rows: rows, Cols: c.cols, PadSize: c.padSize}
	cmd := component.UpdateCmd{Type: cmdSetPadGridConfig, Data: payload}
	c.Component.SendUpdate(cmd)
	return c
}

func (c *PadGridComponent) SetCols(cols int) *PadGridComponent {
	payload := PadGridConfigPayload{Rows: c.rows, Cols: cols, PadSize: c.padSize}
	cmd := component.UpdateCmd{Type: cmdSetPadGridConfig, Data: payload}
	c.Component.SendUpdate(cmd)
	return c
}

func (c *PadGridComponent) SetPadSize(size int) *PadGridComponent {
	payload := PadGridConfigPayload{Rows: c.rows, Cols: c.cols, PadSize: size}
	cmd := component.UpdateCmd{Type: cmdSetPadGridConfig, Data: payload}
	c.Component.SendUpdate(cmd)
	return c
}

func (c *PadGridComponent) SetPreset(preset *preset.Preset) *PadGridComponent {
	cmd := component.UpdateCmd{Type: cmdSetPadGridPreset, Data: preset}
	c.Component.SendUpdate(cmd)
	return c
}

// Destroy cleans up any subscriptions before removal
func (c *PadGridComponent) Destroy() {
	eventbus.Bus.Unsubscribe(events.ComponentClickEventKey, c.UUID())
	for _, p := range c.pads {
		p.Destroy()
	}
	c.pads = nil
	c.Component.Destroy()
}

func (c *PadGridComponent) Layout() {
	c.drainEvents()

	c.Component.ProcessUpdates()

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: 0, Y: 0})
	defer imgui.PopStyleVarV(2)

	selected := c.selectedPad
	cols := c.cols
	pads := c.pads

	for i, p := range pads {
		isSelected := selected != nil && p == selected
		p.SetSelected(isSelected)
		p.Build()
		if cols > 0 && i%cols != cols-1 {
			imgui.SameLine()
		}
	}

	// Only publish event if the selected pad has changed
	if selected != nil && selected != c.lastPublishedPadPtr {
		eventbus.Bus.Publish(events.PadGridEventRecord{
			EventType: events.PadGridSelectEvent,
			Pad:       selected,
			OwnerID:   c.ownerID,
		})
		c.lastPublishedPadPtr = selected
	}

}

// SetOwnerID sets the UUID of the owning window for event filtering
func (c *PadGridComponent) SetOwnerID(ownerID string) {
	c.ownerID = ownerID
}
