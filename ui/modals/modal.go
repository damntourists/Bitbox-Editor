package modals

import (
	"bitbox-editor/ui/events"
	"bitbox-editor/ui/fonts"
	"bitbox-editor/ui/types"
	"context"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

type (
	Modal struct {
		noClose  bool
		open     bool
		title    string
		subtitle string
		icon     string

		config       *ModalConfig
		WindowEvents signals.Signal[events.WindowEventRecord]

		layoutBuilder types.WindowLayoutBuilder
	}
)

func (m *Modal) Title() string { return fmt.Sprintf("%s %s", fonts.Icon(m.icon), m.title) }
func (m *Modal) Icon() string  { return m.icon }

func (m *Modal) Open() {
	m.open = true
	//imgui.OpenPopupStr(m.Title())

	m.WindowEvents.Emit(context.Background(), events.WindowEventRecord{
		Type:        events.WindowOpen,
		WindowTitle: m.Title(),
	})
}

func (m *Modal) Close() {
	m.open = false
	imgui.CloseCurrentPopup()
	m.WindowEvents.Emit(context.Background(), events.WindowEventRecord{
		Type:        events.WindowClose,
		WindowTitle: m.Title(),
	})
}

func (m *Modal) IsOpen() bool { return m.open }

func (m *Modal) Style() func() {
	return func() {}
}

func (m *Modal) Build() {
	if !m.open {
		return
		//
	} else {
		imgui.OpenPopupStr(m.Title())
	}

	var styleFin func()
	if m.layoutBuilder != nil {
		styleFin = m.layoutBuilder.Style()
	} else {
		styleFin = m.Style()
	}
	defer styleFin()

	if imgui.BeginPopupModalV(m.Title(), &m.open, m.config.Combined()) {
		if m.layoutBuilder != nil {
			m.layoutBuilder.Layout()
		} else {
			m.Layout()
		}
		imgui.EndPopup()
	}
}

func (m *Modal) Layout() { panic("Layout not implemented. Please check that layoutBuilder is set.") }

func NewModal(title string, icon string, config *ModalConfig) *Modal {
	if config == nil {
		config = NewModalConfig()
	}

	return &Modal{
		title:        title,
		icon:         icon,
		noClose:      false,
		open:         false,
		config:       config,
		WindowEvents: signals.New[events.WindowEventRecord](),
	}
}
