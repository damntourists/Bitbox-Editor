package flags

import "github.com/AllenDang/cimgui-go/imgui"

type TreeNodeFlags imgui.TreeNodeFlags

const (
	TreeNodeFlagsNone                 = TreeNodeFlags(imgui.TreeNodeFlagsNone)
	TreeNodeFlagsSelected             = TreeNodeFlags(imgui.TreeNodeFlagsSelected)
	TreeNodeFlagsFramed               = TreeNodeFlags(imgui.TreeNodeFlagsFramed)
	TreeNodeFlagsAllowItemOverlap     = TreeNodeFlags(imgui.TreeNodeFlagsAllowOverlap)
	TreeNodeFlagsNoTreePushOnOpen     = TreeNodeFlags(imgui.TreeNodeFlagsNoTreePushOnOpen)
	TreeNodeFlagsNoAutoOpenOnLog      = TreeNodeFlags(imgui.TreeNodeFlagsNoAutoOpenOnLog)
	TreeNodeFlagsDefaultOpen          = TreeNodeFlags(imgui.TreeNodeFlagsDefaultOpen)
	TreeNodeFlagsOpenOnDoubleClick    = TreeNodeFlags(imgui.TreeNodeFlagsOpenOnDoubleClick)
	TreeNodeFlagsOpenOnArrow          = TreeNodeFlags(imgui.TreeNodeFlagsOpenOnArrow)
	TreeNodeFlagsLeaf                 = TreeNodeFlags(imgui.TreeNodeFlagsLeaf)
	TreeNodeFlagsBullet               = TreeNodeFlags(imgui.TreeNodeFlagsBullet)
	TreeNodeFlagsFramePadding         = TreeNodeFlags(imgui.TreeNodeFlagsFramePadding)
	TreeNodeFlagsSpanAvailWidth       = TreeNodeFlags(imgui.TreeNodeFlagsSpanAvailWidth)
	TreeNodeFlagsSpanFullWidth        = TreeNodeFlags(imgui.TreeNodeFlagsSpanFullWidth)
	TreeNodeFlagsSpanAllColumns       = TreeNodeFlags(imgui.TreeNodeFlagsSpanAllColumns)
	TreeNodeFlagsNavLeftJumpsBackHere = TreeNodeFlags(imgui.TreeNodeFlagsNavLeftJumpsBackHere)
	TreeNodeFlagsCollapsingHeader     = TreeNodeFlags(imgui.TreeNodeFlagsCollapsingHeader)
)
