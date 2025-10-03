package component

import (
	"bitbox-editor/ui/events"
	"bitbox-editor/ui/fonts"
	"context"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

type ComboBoxComponent struct {
	*Component

	label    string
	preview  string
	items    []string
	selected int32

	width float32
	flags imgui.ComboFlags

	Events signals.Signal[events.ComboBoxEventRecord]
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

func (c *ComboBoxComponent) Width() float32 {
	return c.width
}

func (c *ComboBoxComponent) Flags() imgui.ComboFlags {
	return c.flags
}
func (c *ComboBoxComponent) SetFlags(flags imgui.ComboFlags) *ComboBoxComponent {
	c.flags = flags
	return c
}

func (c *ComboBoxComponent) SetSelected(selected int32) *ComboBoxComponent {
	c.selected = selected
	return c
}

func (c *ComboBoxComponent) SetItems(items []string) *ComboBoxComponent {
	c.items = items
	return c
}

func (c *ComboBoxComponent) SetPreview(preview string) *ComboBoxComponent {
	c.preview = preview
	return c
}

func (c *ComboBoxComponent) SetWidth(width float32) *ComboBoxComponent {
	c.width = width
	return c
}

func (c *ComboBoxComponent) Menu() {}

func (c *ComboBoxComponent) Layout() {
	if c.width > 0 {
		imgui.PushItemWidth(c.width)
		defer imgui.PopItemWidth()
	}

	if imgui.BeginComboV(c.label, c.preview, c.flags) {
		for i, item := range c.items {
			if imgui.SelectableBool(fmt.Sprintf("%s##%d", item, i)) {
				c.selected = int32(i)
				c.Events.Emit(context.Background(), events.ComboBoxEventRecord{
					Type: events.SelectionChanged,
					Data: item,
				})
			}

		}
		imgui.EndCombo()
	}

	if int(c.selected) >= 0 {
		c.SetPreview(fmt.Sprintf("%s %s", fonts.Icon("Grid3x2"), c.items[c.selected]))
	}

}

func NewComboBoxComponent(id imgui.ID, label string) *ComboBoxComponent {
	cmb := &ComboBoxComponent{
		Component: NewComponent(id),
		label:     label,
		selected:  0,
		Events:    signals.New[events.ComboBoxEventRecord](),
	}

	return cmb
}
