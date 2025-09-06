package modals

import "github.com/AllenDang/cimgui-go/imgui"

type ModalConfig struct {
	noResize         bool
	alwaysAutoResize bool
	noMove           bool
	noTitleBar       bool
}

func (c *ModalConfig) SetNoResize(value bool) *ModalConfig {
	c.noResize = value
	return c
}

func (c *ModalConfig) SetAlwaysAutoResize(value bool) *ModalConfig {
	c.alwaysAutoResize = value
	return c
}

func (c *ModalConfig) SetNoMove(value bool) *ModalConfig {
	c.noMove = value
	return c
}

func (c *ModalConfig) SetNoTitleBar(value bool) *ModalConfig {
	c.noTitleBar = value
	return c
}

func (c *ModalConfig) Combined() imgui.WindowFlags {
	flags := imgui.WindowFlagsNone | imgui.WindowFlagsPopup
	if c.alwaysAutoResize {
		flags |= imgui.WindowFlagsAlwaysAutoResize
	}
	if c.noResize {
		flags |= imgui.WindowFlagsNoResize
	}
	if c.noMove {
		flags |= imgui.WindowFlagsNoMove
	}
	if c.noTitleBar {
		flags |= imgui.WindowFlagsNoTitleBar
	}

	return flags
}

func NewModalConfig() *ModalConfig {
	return &ModalConfig{
		noResize:         false,
		alwaysAutoResize: true,
		noMove:           true,
		noTitleBar:       false,
	}
}
