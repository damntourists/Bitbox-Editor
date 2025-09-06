package windows

import "github.com/AllenDang/cimgui-go/imgui"

type (
	// WindowConfig holds various configuration flags for a windows
	WindowConfig struct {
		noTitlebar            bool
		noScrollbar           bool
		noMenu                bool
		noMove                bool
		noResize              bool
		noCollapse            bool
		noNav                 bool
		noBackground          bool
		noBringToFront        bool
		noHorizontalScrollbar bool
		noDocking             bool
		noMouseScrolling      bool
	}
)

func (c *WindowConfig) SetNoTitlebar(value bool) *WindowConfig {
	c.noTitlebar = value
	return c
}
func (c *WindowConfig) SetNoScrollbar(value bool) *WindowConfig {
	c.noScrollbar = value
	return c
}
func (c *WindowConfig) SetNoMenu(value bool) *WindowConfig {
	c.noMenu = value
	return c
}
func (c *WindowConfig) SetNoMove(value bool) *WindowConfig {
	c.noMove = value
	return c
}
func (c *WindowConfig) SetNoResize(value bool) *WindowConfig {
	c.noResize = value
	return c
}
func (c *WindowConfig) SetNoCollapse(value bool) *WindowConfig {
	c.noCollapse = value
	return c
}
func (c *WindowConfig) SetNoNav(value bool) *WindowConfig {
	c.noNav = value
	return c
}
func (c *WindowConfig) SetNoBackground(value bool) *WindowConfig {
	c.noBackground = value
	return c
}
func (c *WindowConfig) SetNoBringToFront(value bool) *WindowConfig {
	c.noBringToFront = value
	return c
}
func (c *WindowConfig) SetNoHorizontalScrollbar(value bool) *WindowConfig {
	c.noHorizontalScrollbar = value
	return c
}
func (c *WindowConfig) SetNoDocking(value bool) *WindowConfig {
	c.noDocking = value
	return c
}
func (c *WindowConfig) SetNoMouseScrolling(value bool) *WindowConfig {
	c.noMouseScrolling = value
	return c
}

// Combined returns the combined imgui flags based on the *WindowConfig settings
func (c *WindowConfig) Combined() imgui.WindowFlags {
	flags := imgui.WindowFlagsNone
	if c.noTitlebar {
		flags |= imgui.WindowFlagsNoTitleBar
	}
	if c.noScrollbar {
		flags |= imgui.WindowFlagsNoScrollbar
	}
	if !c.noMenu {
		flags |= imgui.WindowFlagsMenuBar
	}
	if c.noMove {
		flags |= imgui.WindowFlagsNoMove
	}
	if c.noResize {
		flags |= imgui.WindowFlagsNoResize
	}
	if c.noCollapse {
		flags |= imgui.WindowFlagsNoCollapse
	}
	if c.noNav {
		flags |= imgui.WindowFlagsNoNav
	}
	if c.noBackground {
		flags |= imgui.WindowFlagsNoBackground
	}
	if c.noBringToFront {
		flags |= imgui.WindowFlagsNoBringToFrontOnFocus
	}
	if !c.noHorizontalScrollbar {
		flags |= imgui.WindowFlagsHorizontalScrollbar
	}
	if c.noDocking {
		flags |= imgui.WindowFlagsNoDocking
	}
	if c.noMouseScrolling {
		flags |= imgui.WindowFlagsNoScrollWithMouse
	}

	return flags
}

func NewWindowConfig() *WindowConfig {
	return &WindowConfig{
		noTitlebar:            false,
		noScrollbar:           false,
		noMenu:                false,
		noMove:                false,
		noResize:              false,
		noCollapse:            false,
		noNav:                 false,
		noBackground:          false,
		noBringToFront:        false,
		noHorizontalScrollbar: false,
		noDocking:             false,
		noMouseScrolling:      false,
	}
}
