package bitbox

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Document struct {
	XMLName xml.Name `xml:"document"`
	Session Session  `xml:"session"`
}

type Session struct {
	Cells []Cell `xml:"cell"`
}

type SampleParams struct {
	GainDB          int `xml:"gaindb,attr,omitempty"`
	Pitch           int `xml:"pitch,attr,omitempty"`
	PanPos          int `xml:"panpos,attr,omitempty"`
	SamTrigType     int `xml:"samtrigtype,attr,omitempty"`
	LoopMode        int `xml:"loopmode,attr,omitempty"`
	LoopModes       int `xml:"loopmodes,attr,omitempty"`
	MidiMode        int `xml:"midimode,attr,omitempty"`
	MidiOutChan     int `xml:"midioutchan,attr,omitempty"`
	Reverse         int `xml:"reverse,attr,omitempty"`
	CellMode        int `xml:"cellmode,attr,omitempty"`
	EnvAttack       int `xml:"envattack,attr,omitempty"`
	EnvDecay        int `xml:"envdecay,attr,omitempty"`
	EnvSus          int `xml:"envsus,attr,omitempty"`
	EnvRel          int `xml:"envrel,attr,omitempty"`
	SamStart        int `xml:"samstart,attr,omitempty"`
	SamLen          int `xml:"samlen,attr,omitempty"`
	LoopStart       int `xml:"loopstart,attr,omitempty"`
	LoopEnd         int `xml:"loopend,attr,omitempty"`
	QuantSize       int `xml:"quantsize,attr,omitempty"`
	SyncType        int `xml:"synctype,attr,omitempty"`
	ActSlice        int `xml:"actslice,attr,omitempty"`
	OutputBus       int `xml:"outputbus,attr,omitempty"`
	MonoMode        int `xml:"monomode,attr,omitempty"`
	SliceStepMode   int `xml:"slicestepmode,attr,omitempty"`
	ChokeGrp        int `xml:"chokegrp,attr,omitempty"`
	DualFilCutoff   int `xml:"dualfilcutoff,attr,omitempty"`
	RootNote        int `xml:"rootnote,attr,omitempty"`
	BeatCount       int `xml:"beatcount,attr,omitempty"`
	Fx1Send         int `xml:"fx1send,attr,omitempty"`
	Fx2Send         int `xml:"fx2send,attr,omitempty"`
	MultiSamMode    int `xml:"multisammode,attr,omitempty"`
	InterpQual      int `xml:"interpqual,attr,omitempty"`
	PlayThru        int `xml:"playthru,attr,omitempty"`
	SlicerQuantSize int `xml:"slicerquantsize,attr,omitempty"`
	SlicerSync      int `xml:"slicersync,attr,omitempty"`
	PadNote         int `xml:"padnote,attr,omitempty"`
	LoopFadeAmt     int `xml:"loopfadeamt,attr,omitempty"`
	GrainSize       int `xml:"grainsize,attr,omitempty"`
	GrainCount      int `xml:"graincount,attr,omitempty"`
	GainSpreadTen   int `xml:"gainspreadten,attr,omitempty"`
	GrainReadSpeed  int `xml:"grainreadspeed,attr,omitempty"`
	RecPresetLen    int `xml:"recpresetlen,attr,omitempty"`
	RecQuant        int `xml:"recquant,attr,omitempty"`
	RecInput        int `xml:"recinput,attr,omitempty"`
	RecUseThres     int `xml:"recusethres,attr,omitempty"`
	RecThresh       int `xml:"recthresh,attr,omitempty"`
	RecMonOutBus    int `xml:"recmonoutbus,attr,omitempty"`
}

type SamTemplateParams struct {
	GainDB       int `xml:"gaindb,attr,omitempty"`
	Pitch        int `xml:"pitch,attr,omitempty"`
	PanPos       int `xml:"panpos,attr,omitempty"`
	SamTrigType  int `xml:"samtrigtype,attr,omitempty"`
	LoopMode     int `xml:"loopmode,attr,omitempty"`
	LoopModes    int `xml:"loopmodes,attr,omitempty"`
	MidiMode     int `xml:"midimode,attr,omitempty"`
	MidiOutChan  int `xml:"midioutchan,attr,omitempty"`
	Reverse      int `xml:"reverse,attr,omitempty"`
	CellMode     int `xml:"cellmode,attr,omitempty"`
	EnvAttack    int `xml:"envattack,attr,omitempty"`
	EnvDecay     int `xml:"envdecay,attr,omitempty"`
	EnvSus       int `xml:"envsus,attr,omitempty"`
	EnvRel       int `xml:"envrel,attr,omitempty"`
	QuantSize    int `xml:"quantsize,attr,omitempty"`
	SyncType     int `xml:"synctype,attr,omitempty"`
	OutputBus    int `xml:"outputbus,attr,omitempty"`
	MonoMode     int `xml:"monomode,attr,omitempty"`
	SliceStep    int `xml:"slicestepmode,attr,omitempty"`
	ChokeGrp     int `xml:"chokegrp,attr,omitempty"`
	DualCutoff   int `xml:"dualfilcutoff,attr,omitempty"`
	RootNote     int `xml:"rootnote,attr,omitempty"`
	BeatCount    int `xml:"beatcount,attr,omitempty"`
	Fx1Send      int `xml:"fx1send,attr,omitempty"`
	Fx2Send      int `xml:"fx2send,attr,omitempty"`
	InterpQual   int `xml:"interpqual,attr,omitempty"`
	PlayThru     int `xml:"playthru,attr,omitempty"`
	DefTemplate  int `xml:"deftemplate,attr,omitempty"`
	RecPresetLen int `xml:"recpresetlen,attr,omitempty"`
	RecQuant     int `xml:"recquant,attr,omitempty"`
	RecInput     int `xml:"recinput,attr,omitempty"`
	RecUseThres  int `xml:"recusethres,attr,omitempty"`
	RecThresh    int `xml:"recthresh,attr,omitempty"`
	RecMonOutBus int `xml:"recmonoutbus,attr,omitempty"`
}

type DelayParams struct {
	DelayMuTime   int `xml:"delaymustime,attr,omitempty"`
	Feedback      int `xml:"feedback,attr,omitempty"`
	DelayBeatSync int `xml:"dealybeatsync,attr,omitempty"`
	Delay         int `xml:"delay,attr,omitempty"`
}

type ReverbParams struct {
	Decay    int `xml:"decay,attr,omitempty"`
	PreDelay int `xml:"predelay,attr,omitempty"`
	Damping  int `xml:"damping,attr,omitempty"`
}

type Cell struct {
	Row      *int   `xml:"row,attr,omitempty"`
	Column   *int   `xml:"column,attr,omitempty"`
	Layer    *int   `xml:"layer,attr,omitempty"`
	Filename string `xml:"filename,attr,omitempty"`
	Name     string `xml:"name,attr,omitempty"`
	Type     string `xml:"type,attr"`
	Params   any    `xml:"-"`

	ModSources []ModSource   `xml:"modsource"`
	Slices     *Slices       `xml:"slices"`
	Sequence   *NoteSequence `xml:"-"` // TODO
}

type BitcrusherParams struct {
	// TODO: bitcrusher params
}

type IOConnectInParams struct {
	InputIOCon string `xml:"inputiocon,attr,omitempty"`
}

type IOConnectOutParams struct {
	OutputIOCon string `xml:"outputiocon,attr,omitempty"`
}

type SongParams struct {
	GlobTempo int `xml:"globtempo,attr,omitempty"`
	SongMode  int `xml:"songmode,attr,omitempty"`
	SectCount int `xml:"sectcount,attr,omitempty"`
	SectLoop  int `xml:"sectloop,attr,omitempty"`
	Swing     int `xml:"swing,attr,omitempty"`
	KeyMode   int `xml:"keymode,attr,omitempty"`
	KeyRoot   int `xml:"keyroot,attr,omitempty"`
}

type FilterParams struct {
	Cutoff     int `xml:"cutoff,attr,omitempty"`
	Res        int `xml:"res,attr,omitempty"`
	FilterType int `xml:"filtertype,attr,omitempty"`
	FxTrigMode int `xml:"fxtrigmode,attr,omitempty"`
}

type NoteseqParams struct {
	NoteStepLen   int `xml:"notesteplen,attr,omitempty"`
	NoteStepCount int `xml:"notestepcount,attr,omitempty"`
	DutyCycle     int `xml:"dutycyc,attr,omitempty"`
	MidiOutChan   int `xml:"midioutchan,attr,omitempty"`
	QuantSize     int `xml:"quantsize,attr,omitempty"`
	PadNote       int `xml:"padnote,attr,omitempty"`
	DispMode      int `xml:"dispmode,attr,omitempty"`
	SeqPlayEnable int `xml:"seqplayenable,attr,omitempty"`
}

type NoteSequence struct {
	// TODO: noteseq params
}
type AssetParams struct {
	RootNote       int `xml:"rootnote,attr,omitempty"`
	KeyRangeBottom int `xml:"keyrangebottom,attr,omitempty"`
	KeyRangeTop    int `xml:"keyrangetop,attr,omitempty"`
	AssSrcRow      int `xml:"asssrcrow,attr,omitempty"`
	AssSrcCol      int `xml:"asssrccol,attr,omitempty"`
}

type SectionParams struct {
	SectionLenBars int `xml:"sectionlenbars,attr,omitempty"`
}
type EqParams struct {
	EqActBand int `xml:"eqactband,attr,omitempty"`

	EqGain   int `xml:"eqgain,attr,omitempty"`
	EqCutoff int `xml:"eqcutoff,attr,omitempty"`
	EqRes    int `xml:"eqres,attr,omitempty"`
	EqEnable int `xml:"eqenable,attr,omitempty"`
	EqType   int `xml:"eqtype,attr,omitempty"`

	EqGain2   int `xml:"eqgain2,attr,omitempty"`
	EqCutoff2 int `xml:"eqcutoff2,attr,omitempty"`
	EqRes2    int `xml:"eqres2,attr,omitempty"`
	EqEnable2 int `xml:"eqenable2,attr,omitempty"`
	EqType2   int `xml:"eqtype2,attr,omitempty"`

	EqGain3   int `xml:"eqgain3,attr,omitempty"`
	EqCutoff3 int `xml:"eqcutoff3,attr,omitempty"`
	EqRes3    int `xml:"eqres3,attr,omitempty"`
	EqEnable3 int `xml:"eqenable3,attr,omitempty"`
	EqType3   int `xml:"eqtype3,attr,omitempty"`

	EqGain4   int `xml:"eqgain4,attr,omitempty"`
	EqCutoff4 int `xml:"eqcutoff4,attr,omitempty"`
	EqRes4    int `xml:"eqres4,attr,omitempty"`
	EqEnable4 int `xml:"eqenable4,attr,omitempty"`
	EqType4   int `xml:"eqtype4,attr,omitempty"`
}
type NullParams struct{}

type ParamSet map[string]string

func newParamsForType(cellType string) (any, error) {
	switch cellType {
	case "sample":
		return &SampleParams{}, nil
	case "samtempl":
		return &SamTemplateParams{}, nil
	case "delay":
		return &DelayParams{}, nil
	case "reverb":
		return &ReverbParams{}, nil
	case "filter":
		return &FilterParams{}, nil
	case "bitcrusher":
		return &BitcrusherParams{}, nil
	case "ioconnectin":
		return &IOConnectInParams{}, nil
	case "ioconnectout":
		return &IOConnectOutParams{}, nil
	case "song":
		return &SongParams{}, nil
	case "noteseq":
		return &NoteseqParams{}, nil
	case "asset":
		return &AssetParams{}, nil
	case "section":
		return &SectionParams{}, nil
	case "eq":
		return &EqParams{}, nil
	case "null":
		return &NullParams{}, nil
	default:
		// Fallback to empty params.
		ps := ParamSet{}
		return &ps, nil
	}
}

type WavFile struct {
	Original string
	Resolved string

	Row, Column, Layer *int
	Type               string
}

func (d *Document) ResolveWavFiles(baseDir string) ([]WavFile, error) {
	var out []WavFile

	if baseDir != "" {
		if abs, err := filepath.Abs(baseDir); err == nil {
			baseDir = abs
		}
	}

	resolve := func(orig string, row, col, layer *int, typ string) WavFile {
		w := WavFile{
			Original: orig,
			Row:      row,
			Column:   col,
			Layer:    layer,
			Type:     typ,
		}
		if orig == "" {
			return w
		}

		normalized := normalizePath(orig)

		if p, ok := tryExists(normalized); ok {
			w.Resolved = p
			return w
		}

		if baseDir != "" {
			trimmed := strings.TrimLeft(normalized, string(filepath.Separator))
			j := filepath.Join(baseDir, trimmed)
			if p, ok := tryExists(j); ok {
				w.Resolved = p
				return w
			}
			j2 := filepath.Join(baseDir, normalized)
			if p, ok := tryExists(j2); ok {
				w.Resolved = p
				return w
			}
		}

		if baseDir != "" {
			base := filepath.Base(normalized)
			if found := findFileByName(baseDir, base, 64_000); found != "" {
				w.Resolved = found
				return w
			}
		}

		return w
	}

	for i := range d.Session.Cells {
		c := &d.Session.Cells[i]
		if c.Filename == "" {
			continue
		}
		if strings.EqualFold(filepath.Ext(c.Filename), ".wav") {
			out = append(out, resolve(c.Filename, c.Row, c.Column, c.Layer, c.Type))
		}
	}

	return out, nil
}

func normalizePath(p string) string {
	p = strings.ReplaceAll(p, "\\", string(filepath.Separator))
	return filepath.Clean(p)
}

func tryExists(p string) (string, bool) {
	if p == "" {
		return "", false
	}
	if !filepath.IsAbs(p) {
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
	}
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p, true
	}
	return "", false
}

func findFileByName(root, basename string, maxFiles int) string {
	if basename == "" {
		return ""
	}
	count := 0
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		count++
		if count > maxFiles {
			return io.EOF
		}
		if strings.EqualFold(d.Name(), basename) {
			if st, err := os.Stat(path); err == nil && !st.IsDir() {
				abs := path
				if a, err := filepath.Abs(path); err == nil {
					abs = a
				}
				return &foundErr{path: abs}
			}
		}
		return nil
	})
	return ""
}

type foundErr struct{ path string }

func (e *foundErr) Error() string { return e.path }

func (c *Cell) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "row":
			var v int
			if _, err := fmt.Sscan(a.Value, &v); err == nil {
				c.Row = &v
			}
		case "column":
			var v int
			if _, err := fmt.Sscan(a.Value, &v); err == nil {
				c.Column = &v
			}
		case "layer":
			var v int
			if _, err := fmt.Sscan(a.Value, &v); err == nil {
				c.Layer = &v
			}
		case "filename":
			c.Filename = a.Value
		case "name":
			c.Name = a.Value
		case "type":
			c.Type = a.Value
		}
	}

	for {
		tok, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "params":
				paramsVal, err := newParamsForType(c.Type)
				if err != nil {
					return err
				}
				if err := d.DecodeElement(paramsVal, &t); err != nil {
					return err
				}
				c.Params = paramsVal

			case "modsource":
				var ms ModSource
				if err := d.DecodeElement(&ms, &t); err != nil {
					return err
				}
				c.ModSources = append(c.ModSources, ms)

			case "slices":
				var sl Slices
				if err := d.DecodeElement(&sl, &t); err != nil {
					return err
				}
				c.Slices = &sl

			case "sequence":
				var seq NoteSequence
				if err := d.DecodeElement(&seq, &t); err != nil {
					return err
				}
				c.Sequence = &seq

			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}

		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
	return nil
}

func (c *Cell) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "cell"
	start.Attr = start.Attr[:0]

	if c.Type != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "type"}, Value: c.Type})
	}
	if c.Row != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "row"}, Value: fmt.Sprint(*c.Row)})
	}
	if c.Column != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "column"}, Value: fmt.Sprint(*c.Column)})
	}
	if c.Layer != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "layer"}, Value: fmt.Sprint(*c.Layer)})
	}
	if c.Filename != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "filename"}, Value: c.Filename})
	}
	if c.Name != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "name"}, Value: c.Name})
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	if c.Params != nil {
		if err := e.EncodeElement(c.Params, xml.StartElement{Name: xml.Name{Local: "params"}}); err != nil {
			return err
		}
	}

	for _, ms := range c.ModSources {
		if err := e.EncodeElement(ms, xml.StartElement{Name: xml.Name{Local: "modsource"}}); err != nil {
			return err
		}
	}
	if c.Slices != nil {
		if err := e.EncodeElement(c.Slices, xml.StartElement{Name: xml.Name{Local: "slices"}}); err != nil {
			return err
		}
	}
	if c.Sequence != nil {
		if err := e.EncodeElement(c.Sequence, xml.StartElement{Name: xml.Name{Local: "sequence"}}); err != nil {
			return err
		}
	}

	if err := e.EncodeToken(start.End()); err != nil {
		return err
	}
	return e.Flush()
}

type ModSource struct {
	Dest   string `xml:"dest,attr"`
	Src    string `xml:"src,attr"`
	Slot   *int   `xml:"slot,attr,omitempty"`
	Amount *int   `xml:"amount,attr,omitempty"`
}

type Slices struct{}
