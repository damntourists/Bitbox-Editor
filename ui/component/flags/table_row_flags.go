package flags

import "github.com/AllenDang/cimgui-go/imgui"

type TableRowFlags imgui.TableRowFlags

const (
	TableRowFlagsNone    = TableRowFlags(imgui.TableRowFlagsNone)
	TableRowFlagsHeaders = TableRowFlags(imgui.TableRowFlagsHeaders)
)
