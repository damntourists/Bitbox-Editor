package pad_config

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/component/pad"
	"bitbox-editor/internal/app/component/table"
	"bitbox-editor/internal/app/component/text"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/logging"
	"bitbox-editor/internal/parsing/bitbox"
	"bitbox-editor/internal/preset"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/AllenDang/cimgui-go/imgui"
)

var log = logging.NewLogger("pad_config")

type PadConfigComponent struct {
	*component.Component[*PadConfigComponent]
	eventbus.EventRouter
	component.CommandRouter

	pad    *pad.PadComponent
	preset *preset.Preset

	table *table.TableComponent
}

func NewPadConfigComponent(id imgui.ID, p *preset.Preset) *PadConfigComponent {
	cmp := &PadConfigComponent{}

	cmp.Component = component.NewComponent[*PadConfigComponent](id)

	cmp.table = table.NewTableComponent(imgui.IDStr(fmt.Sprintf("table-%s", cmp.UUID()))).
		SetNoHeader(false).
		SetFlags(imgui.TableFlagsBorders|
			imgui.TableFlagsResizable|
			imgui.TableFlagsSizingStretchSame|
			imgui.TableFlagsRowBg|
			imgui.TableFlagsScrollY).
		SetColumns(
			table.NewTableColumn("Key"),
			table.NewTableColumn("Value").SetFlags(imgui.TableColumnFlagsWidthStretch),
		).
		SetSize(0, -1)

	if p != nil {
		cmp.SetPreset(p)
	}

	cmp.Component.SetLayoutBuilder(cmp)

	cmp.EventRouter.Init(cmp.UUID())
	cmp.OnEvent(events.PadGridSelectKey, cmp.onPadSelect)

	cmp.CommandRouter.Init(cmp.Component)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetPadConfigPad, cmp.onSetPad)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetPadConfigPreset, cmp.onSetPreset)

	return cmp
}

// onPadSelect handles pad selection events
func (c *PadConfigComponent) onPadSelect(event events.Event) {
	e := event.(events.PadGridSelectEvent)
	c.SendUpdate(component.UpdateCmd{Type: cmdSetPadConfigPad, Data: e.Pad})
}

func (c *PadConfigComponent) onSetPad(pad *pad.PadComponent) {
	c.pad = pad
	// Rebuild the table on pad change
	c.rebuildTableRows()
}

func (c *PadConfigComponent) onSetPreset(preset *preset.Preset) {
	c.preset = preset
	// Rebuild the table on preset change
	c.rebuildTableRows()
}

// rebuildTableRows finds the correct cell and builds the table rows for it
func (c *PadConfigComponent) rebuildTableRows() {
	if c.pad == nil || c.preset == nil || c.table == nil {
		// Send empty list to clear the table
		c.table.SetRows()
		return
	}

	config := c.preset.BitboxConfig()
	if config == nil || config.Session == nil {
		c.table.SetRows()
		return
	}

	// Find the cell corresponding to the selected pad
	var cell *bitbox.Cell
	for i, s := range config.Session.Cells {
		if s.Row != nil && *s.Row == c.pad.Row() && s.Column != nil && *s.Column == c.pad.Col() {
			cell = &config.Session.Cells[i]
			break
		}
	}

	if cell == nil {
		c.table.SetRows()
		return
	}

	rows := make([]*table.TableRowComponent, 0, 10)

	// Helper to add a row
	addRow := func(id, key, value string) {
		rowID := imgui.IDStr(fmt.Sprintf("row-%s-%s", c.pad.UUID(), id))
		rows = append(rows, table.NewTableRow(rowID,
			text.NewText(key),
			text.NewText(value),
		))
	}

	// Add basic cell info
	if cell.Row != nil {
		addRow("row", "Row", fmt.Sprintf("%d", *cell.Row))
	}
	if cell.Column != nil {
		addRow("col", "Column", fmt.Sprintf("%d", *cell.Column))
	}
	if cell.Layer != nil {
		addRow("layer", "Layer", fmt.Sprintf("%d", *cell.Layer))
	}
	addRow("filename", "Filename", cell.Filename)
	addRow("name", "Name", cell.Name)
	addRow("type", "Type", cell.Type)

	if cell.Params != nil {
		paramRows := drawParamRows(cell.Params)
		if len(paramRows) > 0 {
			// Create a nested table component
			nestedTableID := imgui.IDStr(fmt.Sprintf("tbl-params-%s", c.pad.UUID()))
			nestedTable := table.NewTableComponent(nestedTableID).
				SetSize(0, float32(len(paramRows))*25).
				SetFlags(
					imgui.TableFlagsBorders|
						imgui.TableFlagsResizable|
						imgui.TableFlagsSizingStretchSame|
						imgui.TableFlagsRowBg,
				).
				SetColumns(
					table.NewTableColumn("Key").SetFlags(imgui.TableColumnFlagsWidthFixed),
					table.NewTableColumn("Value").SetFlags(imgui.TableColumnFlagsWidthStretch),
				).
				SetRows(paramRows...)

			// Add the nested table as a row in the main table
			rowID := imgui.IDStr(fmt.Sprintf("row-%s-params", c.pad.UUID()))
			rows = append(rows, table.NewTableRow(rowID,
				text.NewText("Params"),
				nestedTable,
			))
		}
	}

	c.table.SetRows(rows...)
}

func (c *PadConfigComponent) SetPad(pad *pad.PadComponent) *PadConfigComponent {
	cmd := component.UpdateCmd{Type: cmdSetPadConfigPad, Data: pad}
	c.SendUpdate(cmd)
	return c
}

func (c *PadConfigComponent) SetPreset(preset *preset.Preset) *PadConfigComponent {
	cmd := component.UpdateCmd{Type: cmdSetPadConfigPreset, Data: preset}
	c.SendUpdate(cmd)
	return c
}

func (c *PadConfigComponent) Layout() {
	c.EventRouter.ProcessEvents()
	c.Component.ProcessUpdates()

	if c.preset == nil {
		imgui.Text("No preset selected")
		return
	}
	if c.pad == nil {
		imgui.Text("No pad selected")
		return
	}

	// Render the child table
	c.table.Build()
}

// Destroy cleans up the component
func (c *PadConfigComponent) Destroy() {
	// Unsubscribe from all events
	c.EventRouter.Destroy()

	// Destroy child components
	if c.table != nil {
		c.table.Destroy()
	}

	// Call base destroy
	c.Component.Destroy()
}

// TODO: Move these helpers

type fieldMeta struct {
	index   int
	display string
}

type typeMeta struct {
	fields []fieldMeta
}

var (
	metaCache   sync.Map
	orderMu     sync.RWMutex
	orderByType = make(map[string][]string)
)

func drawParamRows(p any) []*table.TableRowComponent {
	emptyRow := []*table.TableRowComponent{
		table.NewTableRow(imgui.IDStr("row-nil"), text.NewText(""), text.NewText("")),
	}

	if p == nil {
		return emptyRow
	}

	v := reflect.ValueOf(p)
	t := reflect.TypeOf(p)

	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return emptyRow
		}
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return emptyRow
	}

	meta := getOrBuildTypeMeta(t)

	rows := make([]*table.TableRowComponent, 0, len(meta.fields))
	for _, fm := range meta.fields {
		fv := v.Field(fm.index)
		val := valueToStringFast(fv)

		rowID := imgui.IDStr(fmt.Sprintf("row-%s-%d", t.Name(), fm.index))
		rows = append(rows, table.NewTableRow(rowID,
			text.NewText(fm.display),
			text.NewText(val),
		))
	}
	return rows
}

func getOrBuildTypeMeta(t reflect.Type) *typeMeta {
	if m, ok := metaCache.Load(t); ok {
		return m.(*typeMeta)
	}

	fields := make([]fieldMeta, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		display := xmlBaseName(sf.Tag.Get("xml"))
		if display == "" {
			display = sf.Name
		}
		fields = append(fields, fieldMeta{
			index:   i,
			display: display,
		})
	}

	orderMu.RLock()
	order, hasOrder := orderByType[t.Name()]
	orderMu.RUnlock()
	if hasOrder && len(order) > 0 {
		fields = applyCustomOrder(fields, order)
	}

	m := &typeMeta{fields: fields}
	if old, loaded := metaCache.LoadOrStore(t, m); loaded {
		return old.(*typeMeta)
	}
	return m
}

func xmlBaseName(tag string) string {
	if tag == "" || tag == "-" {
		return ""
	}
	parts := strings.Split(tag, ",")
	return strings.TrimSpace(parts[0])
}

func applyCustomOrder(in []fieldMeta, order []string) []fieldMeta {
	indexByName := make(map[string]int, len(in))
	for i, f := range in {
		indexByName[f.display] = i
	}
	out := make([]fieldMeta, 0, len(in))
	used := make([]bool, len(in))

	for _, name := range order {
		if idx, ok := indexByName[name]; ok && !used[idx] {
			out = append(out, in[idx])
			used[idx] = true
		}
	}
	for i := range in {
		if !used[i] {
			out = append(out, in[i])
		}
	}
	return out
}

func valueToStringFast(v reflect.Value) string {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())
	case reflect.Slice, reflect.Array:
		n := v.Len()
		if n == 0 {
			return "[]"
		}
		return fmt.Sprintf("len=%d", n)
	case reflect.Struct:
		return v.Type().Name()
	case reflect.Map:
		return fmt.Sprintf("map(len=%d)", v.Len())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
