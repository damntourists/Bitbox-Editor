package combobox

/*
╭────────────────────╮
│ Combobox Component │
╰────────────────────╯
*/

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/logging"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

var log = logging.NewLogger("combobox")

type ComboBoxComponent struct {
	*component.Component[*ComboBoxComponent]

	label   string
	preview string
	items   []string

	selected int32

	flags imgui.ComboFlags
}

func NewComboBoxComponent(id imgui.ID, label string) *ComboBoxComponent {
	cmp := &ComboBoxComponent{
		label:    label,
		selected: 0,
		items:    make([]string, 0),
		preview:  "",
	}
	cmp.Component = component.NewComponent[*ComboBoxComponent](id, cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)

	return cmp
}

func (c *ComboBoxComponent) handleUpdate(cmd component.UpdateCmd) {
	if c.Component.HandleGlobalUpdate(cmd) {
		return
	}

	switch cmd.Type {
	case cmdSetComboBoxItems:
		c.items = cmd.Data.([]string)
	case cmdSetComboBoxSelected:
		newIndex := cmd.Data.(int32)
		c.selected = newIndex
		// Update preview text when selection changes
		if newIndex >= 0 && int(newIndex) < len(c.items) {
			c.preview = fmt.Sprintf("%s %s", font.Icon("Grid3x2"), c.items[c.selected])
		}
	case cmdSetComboBoxPreview:
		c.preview = cmd.Data.(string)
	case cmdSetComboBoxFlags:
		c.flags = cmd.Data.(imgui.ComboFlags)
	default:
		log.Warn("ComboBoxComponent unhandled update", zap.String("id", c.IDStr()), zap.Any("cmd", cmd))
	}
}

func (c *ComboBoxComponent) Selected() int32 {
	return c.selected
}

func (c *ComboBoxComponent) Items() []string {
	return c.items
}

func (c *ComboBoxComponent) Preview() string {
	return c.preview
}

func (c *ComboBoxComponent) Flags() imgui.ComboFlags {
	return c.flags
}

func (c *ComboBoxComponent) SetFlags(flags imgui.ComboFlags) *ComboBoxComponent {
	cmd := component.UpdateCmd{Type: cmdSetComboBoxFlags, Data: flags}
	c.Component.SendUpdate(cmd)
	return c
}

func (c *ComboBoxComponent) SetSelected(selected int32) *ComboBoxComponent {
	c.selected = selected
	return c
}

func (c *ComboBoxComponent) SetItems(items []string) *ComboBoxComponent {
	cmd := component.UpdateCmd{Type: cmdSetComboBoxItems, Data: items}
	c.Component.SendUpdate(cmd)
	return c
}

func (c *ComboBoxComponent) SetPreview(preview string) *ComboBoxComponent {
	cmd := component.UpdateCmd{Type: cmdSetComboBoxPreview, Data: preview}
	c.Component.SendUpdate(cmd)
	return c
}

func (c *ComboBoxComponent) Menu() { /* *crickets* */ }

func (c *ComboBoxComponent) Layout() {
	c.Component.ProcessUpdates()

	// Get base component width
	width := c.Component.Width()

	// Get local state
	label := c.label
	preview := c.preview
	flags := c.flags
	items := c.items

	if width > 0 {
		imgui.PushItemWidth(width)
		defer imgui.PopItemWidth()
	}

	if imgui.BeginComboV(label, preview, flags) {
		for i, item := range items {
			if imgui.SelectableBool(fmt.Sprintf("%s##%d", item, i)) {
				c.SendUpdate(component.UpdateCmd{Type: cmdSetComboBoxSelected, Data: int32(i)})

				eventbus.Bus.Publish(events.ComboboxEventRecord{
					EventType: events.ComboboxSelectionChangeEvent,
					UUID:      c.UUID(),
					Selected:  item,
				})
			}

		}
		imgui.EndCombo()
	}
}

// Destroy cleans up the component
func (c *ComboBoxComponent) Destroy() {
	// This component doesn't subscribe to any events
	c.Component.Destroy()
}
