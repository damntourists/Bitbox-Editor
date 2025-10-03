package component

import (
	"bitbox-editor/lib/parsing/bitbox"
	"bitbox-editor/lib/preset"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/AllenDang/cimgui-go/imgui"
)

type PadConfigComponent struct {
	*Component

	pad    *PadComponent
	preset *preset.Preset

	//Events signals.Signal[events.PadConfigEventRecord]
}

func (c *PadConfigComponent) SetPad(pad *PadComponent) {
	c.pad = pad
}

func (c *PadConfigComponent) SetPreset(preset *preset.Preset) {
	c.preset = preset

	println(preset.Name)

}

func (c *PadConfigComponent) Menu() {}

func (c *PadConfigComponent) Layout() {
	if c.preset == nil {
		imgui.Text("No preset selected")
		return
	}

	if imgui.CollapsingHeaderBoolPtrV("Bitbox config", nil, imgui.TreeNodeFlagsFramed) {
		for i, s := range c.preset.BitboxConfig().Session.Cells {
			drawTable(i, s)
		}
	}

	if imgui.CollapsingHeaderBoolPtrV("Ableton config", nil, imgui.TreeNodeFlagsFramed) {
		imgui.Text("test")
	}
}

func drawTable(i int, s bitbox.Cell) {
	imgui.SeparatorText(fmt.Sprintf("Cell-%d", i))
	imgui.BeginChildStrV(
		fmt.Sprintf("%d-%s", i, s.Filename), imgui.Vec2{X: 0, Y: 0},
		imgui.ChildFlagsBorders|imgui.ChildFlagsResizeY,
		imgui.WindowFlagsChildWindow)

	var r, c, l = "unset", "unset", "unset"

	if s.Row != nil {
		r = fmt.Sprintf("%d", *s.Row)
	}
	if s.Column != nil {
		c = fmt.Sprintf("%d", *s.Column)
	}
	if s.Layer != nil {
		l = fmt.Sprintf("%d", *s.Layer)
	}

	NewTableComponent(imgui.IDStr(fmt.Sprintf("tbl:%s", s.Filename))).
		Flags(imgui.TableFlagsBorders|
			imgui.TableFlagsResizable|
			imgui.TableFlagsSizingStretchSame|
			imgui.TableFlagsRowBg|
			imgui.TableFlagsScrollY).
		Columns(
			NewTableColumn("Key"),
			NewTableColumn("Value").Flags(imgui.TableColumnFlagsWidthStretch),
		).
		Rows(
			TableRow(
				imgui.IDStr(fmt.Sprintf("tbl%d:row0:%s", i, s.Filename)),
				Text("Row"),
				Text(r),
			),
			TableRow(
				imgui.IDStr(fmt.Sprintf("tbl%d:row1:%s", i, s.Filename)),
				Text("Column"),
				Text(c),
			),
			TableRow(
				imgui.IDStr(fmt.Sprintf("tbl%d:row2:%s", i, s.Filename)),
				Text("Layer"),
				Text(l),
			),
			TableRow(
				imgui.IDStr(fmt.Sprintf("tbl%d:row3:%s", i, s.Filename)),
				Text("Filename"),
				Text(s.Filename),
			),
			TableRow(
				imgui.IDStr(fmt.Sprintf("tbl%d:row4:%s", i, s.Filename)),
				Text("Name"),
				Text(s.Name),
			),
			TableRow(
				imgui.IDStr(fmt.Sprintf("tbl%d:row5:%s", i, s.Filename)),
				Text("Type"),
				Text(s.Type),
			),
			TableRow(
				imgui.IDStr(fmt.Sprintf("tbl%d:row6:%s", i, s.Filename)),
				Text("Params"),
				NewTableComponent(imgui.IDStr(fmt.Sprintf("tbl%d:params:%s", i, s.Filename))).
					Size(0, 200).
					Flags(
						imgui.TableFlagsBorders|
							imgui.TableFlagsResizable|
							imgui.TableFlagsSizingStretchSame|
							imgui.TableFlagsRowBg|
							imgui.TableFlagsScrollY,
					).
					Columns(
						NewTableColumn("Key").Flags(imgui.TableColumnFlagsWidthFixed),
						NewTableColumn("Value").Flags(imgui.TableColumnFlagsWidthStretch),
					).
					Rows(
						drawParamRows(s.Params)...,
					),
			),
		).
		Layout()
	imgui.EndChild()
	imgui.NewLine()
}

type fieldMeta struct {
	index   int
	display string
}
type typeMeta struct {
	fields []fieldMeta
}

var (
	metaCache sync.Map

	orderMu     sync.RWMutex
	orderByType = map[string][]string{}
)

func RegisterParamOrder(typeName string, order []string) {
	orderMu.Lock()
	orderByType[typeName] = append([]string(nil), order...)
	orderMu.Unlock()
}

func drawParamRows(p any) []*TableRowComponent {
	emptyRow := []*TableRowComponent{
		TableRow(imgui.IDStr("row-nil"), Text(""), Text("")),
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
		// Not a struct - nothing to render
		return emptyRow
	}

	meta := getOrBuildTypeMeta(t)

	rows := make([]*TableRowComponent, 0, len(meta.fields))
	for _, fm := range meta.fields {
		fv := v.Field(fm.index)
		val := valueToStringFast(fv)

		rowID := imgui.IDStr(fmt.Sprintf("row-%s-%d", t.Name(), fm.index))
		rows = append(rows, TableRow(rowID,
			Text(fm.display),
			Text(val),
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
		// Skip unexported
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

func NewPadConfigComponent(id imgui.ID, p *preset.Preset) *PadConfigComponent {
	return &PadConfigComponent{
		Component: NewComponent(id),
		pad:       nil,
		preset:    p,
	}
}
