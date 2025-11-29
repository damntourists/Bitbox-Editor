/*
COMPONENT: COMBOBOX

	    A dropdown selection component. Broadcasts changes via
	    the global EventBus.

	    ┌──────────────────────┐      ┌───────────────────────┐
	    │ ⊞ Preview Text     ▼ │  ─▶  │ ⊞ Preview Text      ▼ │
	    └──────────────────────┘      │┌─────────────────────┐│
	           (Collapsed)            ││ Option A            ││
									  ││ Option B            ││
									  │└─────────────────────┘│
									  └───────────────────────┘
											  (Expanded)
*/
package combobox

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/logging"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
)

var log = logging.NewLogger("combobox")

/*
┌────────────────────────────────────────────────────────────┐
│ ComboBoxComponent                                          │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ *component.Component[*Button]                        │  │
│  │  • Animation Engine                                  │  │
│  │  • Event Bus (interaction events)                    │  │
│  │  • Update Command Queue (thread-safe updates)        │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ component.CommandRouter                              │  │
│  │  • Cmd -> Handler Mapping                            │  │
│  └──────────────────────────────────────────────────────┘  │
│  • Local State: Items[], Selected(int), Preview(string)    │
│  • Output: ComboboxSelectionChangeEvent (Global)           │
└────────────────────────────────────────────────────────────┘
*/
type ComboBoxComponent struct {
	*component.Component[*ComboBoxComponent]
	component.CommandRouter

	label   string
	preview string
	items   []string

	selected int32

	flags imgui.ComboFlags
}

/*
COMMAND ROUTING
╭────────────────────────┬───────────────────────────────────╮
│ Command Type           │ Handler Method                    │
├────────────────────────┼───────────────────────────────────┤
│ cmdSetComboBoxItems    │ onSetItems( []string )            │
│ cmdSetComboBoxSelected │ onSetSelected( int32 )            │
│ cmdSetComboBoxPreview  │ onSetPreview( string )            │
│ cmdSetComboBoxFlags    │ onSetFlags( imgui.ComboFlags )    │
╰────────────────────────┴───────────────────────────────────╯
*/
func NewComboBoxComponent(id imgui.ID, label string) *ComboBoxComponent {
	cmp := &ComboBoxComponent{
		label:    label,
		selected: 0,
		items:    make([]string, 0),
		preview:  "",
	}
	cmp.Component = component.NewComponent[*ComboBoxComponent](id)
	cmp.Component.SetLayoutBuilder(cmp)

	cmp.CommandRouter.Init(cmp.Component)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetComboBoxItems, cmp.onSetItems)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetComboBoxSelected, cmp.onSetSelected)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetComboBoxPreview, cmp.onSetPreview)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetComboBoxFlags, cmp.onSetFlags)

	return cmp
}

func (c *ComboBoxComponent) onSetItems(items []string) {
	c.items = items
}

func (c *ComboBoxComponent) onSetSelected(newIndex int32) {
	c.selected = newIndex
	// Update preview text when selection changes
	if newIndex >= 0 && int(newIndex) < len(c.items) {
		c.preview = fmt.Sprintf("%s %s", font.Icon("Grid3x2"), c.items[c.selected])
	}
}

func (c *ComboBoxComponent) onSetPreview(preview string) {
	c.preview = preview
}

func (c *ComboBoxComponent) onSetFlags(flags imgui.ComboFlags) {
	c.flags = flags
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
	c.SendUpdate(cmd)
	return c
}

func (c *ComboBoxComponent) SetSelected(selected int32) *ComboBoxComponent {
	c.selected = selected
	return c
}

func (c *ComboBoxComponent) SetItems(items []string) *ComboBoxComponent {
	cmd := component.UpdateCmd{Type: cmdSetComboBoxItems, Data: items}
	c.SendUpdate(cmd)
	return c
}

func (c *ComboBoxComponent) SetPreview(preview string) *ComboBoxComponent {
	cmd := component.UpdateCmd{Type: cmdSetComboBoxPreview, Data: preview}
	c.SendUpdate(cmd)
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

				eventbus.Bus.Publish(events.ComboboxSelectionChangeEvent{
					UUID:     c.UUID(),
					Selected: item,
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
