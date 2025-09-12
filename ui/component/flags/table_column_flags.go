package flags

import "github.com/AllenDang/cimgui-go/imgui"

type TableColumnFlags imgui.TableColumnFlags

const (
	TableColumnFlagsNone                 = TableColumnFlags(imgui.TableColumnFlagsNone)
	TableColumnFlagsDefaultHide          = TableColumnFlags(imgui.TableColumnFlagsDefaultHide)
	TableColumnFlagsDefaultSort          = TableColumnFlags(imgui.TableColumnFlagsDefaultSort)
	TableColumnFlagsWidthStretch         = TableColumnFlags(imgui.TableColumnFlagsWidthStretch)
	TableColumnFlagsWidthFixed           = TableColumnFlags(imgui.TableColumnFlagsWidthFixed)
	TableColumnFlagsNoResize             = TableColumnFlags(imgui.TableColumnFlagsNoResize)
	TableColumnFlagsNoReorder            = TableColumnFlags(imgui.TableColumnFlagsNoReorder)
	TableColumnFlagsNoHide               = TableColumnFlags(imgui.TableColumnFlagsNoHide)
	TableColumnFlagsNoClip               = TableColumnFlags(imgui.TableColumnFlagsNoClip)
	TableColumnFlagsNoSort               = TableColumnFlags(imgui.TableColumnFlagsNoSort)
	TableColumnFlagsNoSortAscending      = TableColumnFlags(imgui.TableColumnFlagsNoSortAscending)
	TableColumnFlagsNoSortDescending     = TableColumnFlags(imgui.TableColumnFlagsNoSortDescending)
	TableColumnFlagsNoHeaderWidth        = TableColumnFlags(imgui.TableColumnFlagsNoHeaderWidth)
	TableColumnFlagsPreferSortAscending  = TableColumnFlags(imgui.TableColumnFlagsPreferSortAscending)
	TableColumnFlagsPreferSortDescending = TableColumnFlags(imgui.TableColumnFlagsPreferSortDescending)
	TableColumnFlagsIndentEnable         = TableColumnFlags(imgui.TableColumnFlagsIndentEnable)
	TableColumnFlagsIndentDisable        = TableColumnFlags(imgui.TableColumnFlagsIndentDisable)

	TableColumnFlagsIsEnabled = TableColumnFlags(imgui.TableColumnFlagsIsEnabled)
	TableColumnFlagsIsVisible = TableColumnFlags(imgui.TableColumnFlagsIsVisible)
	TableColumnFlagsIsSorted  = TableColumnFlags(imgui.TableColumnFlagsIsSorted)
	TableColumnFlagsIsHovered = TableColumnFlags(imgui.TableColumnFlagsIsHovered)

	TableColumnFlagsWidthMask      = TableColumnFlags(imgui.TableColumnFlagsWidthMask)
	TableColumnFlagsIndentMask     = TableColumnFlags(imgui.TableColumnFlagsIndentMask)
	TableColumnFlagsStatusMask     = TableColumnFlags(imgui.TableColumnFlagsStatusMask)
	TableColumnFlagsNoDirectResize = TableColumnFlags(imgui.TableColumnFlagsNoDirectResize)
)
