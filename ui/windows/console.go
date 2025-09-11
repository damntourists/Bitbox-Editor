package windows

import (
	"bitbox-editor/lib/events"
	"bitbox-editor/lib/logging"
	"bitbox-editor/ui/fonts"
	"bitbox-editor/ui/theme"
	"context"
	"sort"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

var (
	logs []events.LogRecord
)

type ConsoleWindow struct {
	*Window

	tableId    string
	tableFlags imgui.TableFlags

	scrollToBottom bool
}

func (w *ConsoleWindow) onLogEvent(ctx context.Context, e events.LogRecord) {
	logs = append(logs, e)
	sort.Slice(logs, func(i, j int) bool {
		timeI, errI := time.Parse(time.RFC3339, logs[i].Timestamp)
		if errI != nil {
			return false // Handle invalid timestamp formats gracefully
		}
		timeJ, errJ := time.Parse(time.RFC3339, logs[j].Timestamp)
		if errJ != nil {
			return true // Handle invalid timestamp formats gracefully
		}
		return timeI.Before(timeJ)
	})
}

func (w *ConsoleWindow) Menu() {}

func (w *ConsoleWindow) Style() func() {
	imgui.PushStyleColorVec4(imgui.ColWindowBg, theme.GetCurrentTheme().Style.Colors.ChildBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColChildBg, theme.GetCurrentTheme().Style.Colors.ChildBg.Vec4)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 2, Y: 2})

	return func() {
		imgui.PopStyleVar()
		imgui.PopStyleColorV(2)
	}
}

func (w *ConsoleWindow) Layout() {
	if imgui.BeginTableV(w.tableId, 4, w.tableFlags, imgui.Vec2{}, 0) {
		stretch := imgui.TableColumnFlagsWidthStretch
		static := imgui.TableColumnFlagsWidthFixed
		imgui.TableSetupColumnV("level", static, 32, 0)
		imgui.TableSetupColumnV("name", static, 48, 0)
		imgui.TableSetupColumnV("caller", static, 250, 0)
		imgui.TableSetupColumnV("msg", stretch, 1, 0)

		clipper := imgui.NewListClipper()
		defer clipper.Destroy()

		clipper.Begin(int32(len(logs)))

		for clipper.Step() {
			for i := clipper.DisplayStart(); i < clipper.DisplayEnd(); i++ {
				log := logs[i]
				imgui.TableNextRowV(imgui.TableRowFlags(0), 0)
				imgui.TableSetBgColorV(
					imgui.TableBgTargetRowBg0,
					imgui.ColorU32Vec4(theme.GetCurrentTheme().Style.Colors.ChildBg.Vec4),
					-1,
				)
				imgui.PushFont(fonts.FontCode)
				imgui.TableNextColumn()
				imgui.Text(log.Level)
				imgui.TableNextColumn()
				imgui.Text(log.Name)
				imgui.TableNextColumn()
				imgui.Text(log.Caller)
				imgui.TableNextColumn()
				imgui.Text(log.Message)
				imgui.PopFont()
			}
		}
		clipper.End()

		if w.scrollToBottom {
			imgui.SetScrollHereYV(1)
		}

		if imgui.ScrollY() == imgui.ScrollMaxY() {
			w.scrollToBottom = true
		} else {
			w.scrollToBottom = false
		}

		// TODO: Only stick to bottom if scrolled to bottom.
		//imgui.SetScrollYFloat(imgui.ScrollMaxY())
		imgui.EndTable()

	}
}

func NewConsoleWindow() *ConsoleWindow {
	w := &ConsoleWindow{
		Window:  NewWindow("Console", "SquareTerminal", NewWindowConfig()),
		tableId: "console",
		tableFlags: imgui.TableFlagsResizable |
			imgui.TableFlagsNoBordersInBodyUntilResize |
			imgui.TableFlagsScrollY |
			imgui.TableFlagsNoPadOuterX |
			imgui.TableFlagsRowBg | imgui.TableFlagsSizingFixedFit,
		scrollToBottom: true,
	}

	w.Window.layoutBuilder = w

	logging.LogEvent.AddListener(w.onLogEvent, "logEvents")
	return w
}
