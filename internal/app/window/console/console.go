package console

/*
┍━━━━━━━━━━━━━━━╳┑
│ Console Window │
└────────────────┘
*/

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/app/window"
	"bitbox-editor/internal/config"
	"bitbox-editor/internal/logging"
	"fmt"
	"sort"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

var log = logging.NewLogger("console")

const maxLogsPerFrame = 100

type logEntryWithHeight struct {
	record  logging.LogRecord
	numRows int
}

type ConsoleWindow struct {
	*window.Window[*ConsoleWindow]
	component.CommandRouter

	logs           []logging.LogRecord
	logEntries     []logEntryWithHeight
	tableId        string
	tableFlags     imgui.TableFlags
	scrollToBottom bool
	needsSort      bool
	needsRecalc    bool
}

func NewConsoleWindow() *ConsoleWindow {
	w := &ConsoleWindow{
		logs:       make([]logging.LogRecord, 0, 100),
		logEntries: make([]logEntryWithHeight, 0, 100),
		tableId:    "console",
		tableFlags: imgui.TableFlagsResizable |
			imgui.TableFlagsNoBordersInBodyUntilResize |
			imgui.TableFlagsScrollY |
			imgui.TableFlagsNoPadOuterX |
			imgui.TableFlagsSizingFixedFit,
		scrollToBottom: true,
		needsSort:      true,
		needsRecalc:    true,
	}

	w.Window = window.NewWindow[*ConsoleWindow]("Console", "SquareTerminal")

	w.Window.SetLayoutBuilder(w)

	w.CommandRouter.Init(w.Window)
	component.OnCommandTyped(&w.CommandRouter, cmdConsoleAddLog, w.onAddLog)

	return w
}

func (w *ConsoleWindow) onAddLog(record logging.LogRecord) {
	w.logs = append(w.logs, record)

	// Trim logs if they exceed max lines
	maxLines := config.GetConsoleMaxLines()
	if len(w.logs) > maxLines {
		w.logs = w.logs[len(w.logs)-maxLines:]
	}

	w.needsSort = true
	w.needsRecalc = true
	w.scrollToBottom = true
}

// drainLogChannel - drains log entries from the global log channel and sends them as
// local update commands to this window
func (w *ConsoleWindow) drainLogChannel() {
	for i := 0; i < maxLogsPerFrame; i++ {
		select {
		case logEntry := <-logging.LogChannel:
			// Create a command with our local command type
			cmd := UpdateCmd{Type: cmdConsoleAddLog, Data: logEntry}
			w.Window.SendUpdate(cmd)
		default:
			return
		}
	}
}

// calculateLogEntries pre-calculates the number of rows each log entry will take
func (w *ConsoleWindow) calculateLogEntries() {
	w.logEntries = make([]logEntryWithHeight, len(w.logs))
	for i, logRecord := range w.logs {
		numRows := 1
		if len(logRecord.Params) > 0 {
			numRows += len(logRecord.Params)
		}
		w.logEntries[i] = logEntryWithHeight{
			record:  logRecord,
			numRows: numRows,
		}
	}
}

func (w *ConsoleWindow) Menu() {}

func (w *ConsoleWindow) Style() func() {
	imgui.PushStyleColorVec4(imgui.ColWindowBg, theme.GetCurrentTheme().Style.Colors.ScrollbarBg.Vec4)
	imgui.PushStyleColorVec4(imgui.ColChildBg, theme.GetCurrentTheme().Style.Colors.ScrollbarBg.Vec4)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 2, Y: 2})

	return func() {
		imgui.PopStyleVar()
		imgui.PopStyleColorV(2)
	}
}

func (w *ConsoleWindow) Layout() {
	w.drainLogChannel()
	w.Window.ProcessUpdates()

	if w.needsSort {
		sort.SliceStable(w.logs, func(i, j int) bool {
			timeI, errI := time.Parse(time.RFC3339, w.logs[i].Timestamp)
			timeJ, errJ := time.Parse(time.RFC3339, w.logs[j].Timestamp)
			if errI != nil && errJ != nil {
				return false
			}
			if errI != nil {
				return false
			}
			if errJ != nil {
				return true
			}
			return timeI.Before(timeJ)
		})
		w.needsSort = false
	}

	// Recalculate log entries with row counts if needed
	if w.needsRecalc {
		w.calculateLogEntries()
		w.needsRecalc = false
	}

	tableId := w.tableId
	tableFlags := w.tableFlags
	scrollToBottom := w.scrollToBottom

	// TODO: Derive from colormap
	levelColors := map[string]imgui.Vec4{
		"debug": {X: 0.5, Y: 0.5, Z: 0.5, W: 1.0},
		"info":  {X: 1.0, Y: 1.0, Z: 1.0, W: 1.0},
		"warn":  {X: 1.0, Y: 1.0, Z: 0.0, W: 1.0},
		"error": {X: 1.0, Y: 0.0, Z: 0.0, W: 1.0},
	}

	if imgui.BeginTableV(tableId, 4, tableFlags, imgui.Vec2{}, 0) {
		defer imgui.EndTable()

		stretch := imgui.TableColumnFlagsWidthStretch
		static := imgui.TableColumnFlagsWidthFixed
		imgui.TableSetupColumnV("level", static, 32, 0)
		imgui.TableSetupColumnV("name", static, 48, 0)
		imgui.TableSetupColumnV("caller", static, 250, 0)
		imgui.TableSetupColumnV("msg", stretch, 1, 0)

		// Render all log entries
		for _, entry := range w.logEntries {
			logEntry := entry.record

			imgui.PushFont(font.FontCode, 0)
			imgui.PushStyleColorVec4(imgui.ColText, levelColors[logEntry.Level])

			// Main log row
			imgui.TableNextRow()
			imgui.TableNextColumn()
			imgui.Text(logEntry.Level)
			imgui.TableNextColumn()
			imgui.Text(logEntry.Name)
			imgui.TableNextColumn()
			imgui.Text(logEntry.Caller)
			imgui.TableNextColumn()
			imgui.Text(logEntry.Message)

			// Params as sub-rows
			if len(logEntry.Params) > 0 {
				// Sort the keys for consistent ordering
				keys := make([]string, 0, len(logEntry.Params))
				for key := range logEntry.Params {
					keys = append(keys, key)
				}
				sort.Strings(keys)

				for _, key := range keys {
					imgui.TableNextRow()
					imgui.TableNextColumn()
					imgui.TableNextColumn()
					imgui.TableNextColumn()
					imgui.TableNextColumn()
					textCol := levelColors[logEntry.Level]
					textCol.W *= 0.75
					imgui.PushStyleColorVec4(imgui.ColText, textCol)
					imgui.Text(fmt.Sprintf("%s: %v", key, logEntry.Params[key]))
					imgui.PopStyleColor()
				}
			}

			imgui.PopStyleColor()
			imgui.PopFont()
		}

		if scrollToBottom {
			imgui.SetScrollHereYV(1)
			w.scrollToBottom = false
		}
	}
}

// Destroy cleans up the component
func (w *ConsoleWindow) Destroy() {
	// Console doesn't subscribe to any events, so just call the base destroy method.
	w.Window.Destroy()
}
