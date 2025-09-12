package ableton

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

// Ableton Live project
type Ableton struct {
	XMLName           xml.Name `xml:"Ableton"`
	MajorVersion      string   `xml:"MajorVersion,attr"`
	MinorVersion      string   `xml:"MinorVersion,attr"`
	SchemaChangeCount int      `xml:"SchemaChangeCount,attr"`
	Creator           string   `xml:"Creator,attr"`
	LiveSet           LiveSet  `xml:"LiveSet"`
}

type LiveSet struct {
	Tracks      Tracks      `xml:"Tracks"`
	MasterTrack MasterTrack `xml:"MasterTrack"`
	SceneNames  SceneNames  `xml:"SceneNames"`
}

type Tracks struct {
	AudioTracks []AudioTrack `xml:"AudioTrack"`
}

type AudioTrack struct {
	Id          int              `xml:"Id,attr"`
	Name        TrackName        `xml:"Name"`
	ColorIndex  AttrInt          `xml:"ColorIndex"`
	DeviceChain TrackDeviceChain `xml:"DeviceChain"`
}

type TrackName struct {
	EffectiveName AttrString `xml:"EffectiveName"`
	UserName      AttrString `xml:"UserName"`
}

type TrackDeviceChain struct {
	MainSequencer   Sequencer         `xml:"MainSequencer"`
	FreezeSequencer Sequencer         `xml:"FreezeSequencer"`
	DeviceChain     *InnerDeviceChain `xml:"DeviceChain"`
}

type InnerDeviceChain struct {
	Devices *struct{} `xml:"Devices"`
}

type Sequencer struct {
	ClipSlotList ClipSlotList `xml:"ClipSlotList"`
}

type ClipSlotList struct {
	ClipSlots []ClipSlot `xml:"ClipSlot"`
}

type ClipSlot struct {
	HasStop AttrBool      `xml:"HasStop"`
	Inner   InnerClipSlot `xml:"ClipSlot"`
}

type InnerClipSlot struct {
	Value ClipValue `xml:"Value"`
}

type ClipValue struct {
	AudioClip *AudioClip `xml:"AudioClip"`
}

type AudioClip struct {
	Time             float64           `xml:"Time,attr"`
	WarpMarkers      WarpMarkers       `xml:"WarpMarkers"`
	MarkersGenerated AttrBool          `xml:"MarkersGenerated"`
	IsWarped         AttrBool          `xml:"IsWarped"`
	SampleRef        SampleRef         `xml:"SampleRef"`
	SampleVolume     AttrFloat64       `xml:"SampleVolume"`
	Loop             ClipLoop          `xml:"Loop"`
	Name             AttrString        `xml:"Name"`
	ColorIndex       AttrInt           `xml:"ColorIndex"`
	TimeSignature    ClipTimeSignature `xml:"TimeSignature"`
}

type WarpMarkers struct {
	WarpMarker []WarpMarker `xml:"WarpMarker"`
}

type WarpMarker struct {
	SecTime  float64 `xml:"SecTime,attr"`
	BeatTime float64 `xml:"BeatTime,attr"`
}

type SampleRef struct {
	FileRef FileRef `xml:"FileRef"`
}

type FileRef struct {
	HasRelativePath  AttrBool     `xml:"HasRelativePath"`
	RelativePathType AttrInt      `xml:"RelativePathType"`
	RelativePath     RelativePath `xml:"RelativePath"`
	Name             AttrString   `xml:"Name"`
}

type RelativePath struct {
	Elements []RelativePathElement `xml:"RelativePathElement"`
}

type RelativePathElement struct {
	Dir string `xml:"Dir,attr"`
}

type ClipLoop struct {
	LoopStart       AttrFloat64 `xml:"LoopStart"`
	LoopEnd         AttrFloat64 `xml:"LoopEnd"`
	StartRelative   AttrFloat64 `xml:"StartRelative"`
	LoopOn          AttrBool    `xml:"LoopOn"`
	OutMarker       AttrFloat64 `xml:"OutMarker"`
	HiddenLoopStart AttrFloat64 `xml:"HiddenLoopStart"`
	HiddenLoopEnd   AttrFloat64 `xml:"HiddenLoopEnd"`
}

type ClipTimeSignature struct {
	TimeSignatures TimeSignatures `xml:"TimeSignatures"`
}

type TimeSignatures struct {
	Remoteable []RemoteableTimeSignature `xml:"RemoteableTimeSignature"`
}

type RemoteableTimeSignature struct {
	Numerator   AttrInt     `xml:"Numerator"`
	Denominator AttrInt     `xml:"Denominator"`
	Time        AttrFloat64 `xml:"Time"`
}

type MasterTrack struct {
	DeviceChain MasterDeviceChain `xml:"DeviceChain"`
}

type MasterDeviceChain struct {
	Mixer MasterMixer `xml:"Mixer"`
}

type MasterMixer struct {
	Tempo         TempoAutomation         `xml:"Tempo"`
	TimeSignature TimeSignatureAutomation `xml:"TimeSignature"`
}

type TempoAutomation struct {
	ArrangerAutomation ArrangerAutomationFloat `xml:"ArrangerAutomation"`
}

type TimeSignatureAutomation struct {
	ArrangerAutomation ArrangerAutomationEnum `xml:"ArrangerAutomation"`
}

type ArrangerAutomationFloat struct {
	Events FloatEvents `xml:"Events"`
}

type ArrangerAutomationEnum struct {
	Events EnumEvents `xml:"Events"`
}

type FloatEvents struct {
	FloatEvent []FloatEvent `xml:"FloatEvent"`
}

type EnumEvents struct {
	EnumEvent []EnumEvent `xml:"EnumEvent"`
}

type FloatEvent struct {
	Time  float64 `xml:"Time,attr"`
	Value float64 `xml:"Value,attr"`
}

type EnumEvent struct {
	Time  float64 `xml:"Time,attr"`
	Value int     `xml:"Value,attr"`
}

type SceneNames struct {
	Scenes []Scene `xml:"Scene"`
}

type Scene struct{}

type AttrString struct {
	Value string `xml:"Value,attr"`
}

type AttrInt struct {
	Value int `xml:"Value,attr"`
}

type AttrFloat64 struct {
	Value float64 `xml:"Value,attr"`
}

type AttrBool struct {
	Value bool `xml:"Value,attr"`
}

func Unmarshal(data []byte) (*Ableton, error) {
	// gzip magic 0x1f 0x8b
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("gzip reader: %w", err)
		}
		defer gr.Close()
		return UnmarshalReader(gr)
	}
	return UnmarshalReader(bytes.NewReader(data))
}

// UnmarshalReader parses an Ableton XML document.
func UnmarshalReader(r io.Reader) (*Ableton, error) {
	var doc Ableton
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("xml decode: %w", err)
	}
	return &doc, nil
}

// Marshal serializes the Ableton document to XML bytes.
func Marshal(doc *Ableton) ([]byte, error) {
	var buf bytes.Buffer
	if err := marshalToWriter(&buf, doc, false, false); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalIndent serializes the Ableton document.
func MarshalIndent(doc *Ableton, indent bool) ([]byte, error) {
	var buf bytes.Buffer
	if err := marshalToWriter(&buf, doc, indent, false); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalGZIP serializes the Ableton document to gzipped XML bytes.
func MarshalGZIP(doc *Ableton, indent bool) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if err := marshalToWriter(gw, doc, indent, true); err != nil {
		_ = gw.Close()
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}
	return buf.Bytes(), nil
}

func UnmarshalFile(path string) (*Ableton, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	// Check if gzip magic 0x1f 0x8b
	var hdr [2]byte
	n, _ := f.Read(hdr[:])
	if n == 2 && hdr[0] == 0x1f && hdr[1] == 0x8b {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		gr, err := gzip.NewReader(f)
		if err != nil {
			return nil, fmt.Errorf("gzip reader: %w", err)
		}
		defer gr.Close()
		return UnmarshalReader(gr)
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return UnmarshalReader(f)
}

func MarshalFile(path string, doc *Ableton, indent, gzipOut bool) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	var w io.Writer = f
	var gw *gzip.Writer
	if gzipOut {
		gw = gzip.NewWriter(f)
		w = gw
	}

	if err := marshalToWriter(w, doc, indent, gzipOut); err != nil {
		if gw != nil {
			_ = gw.Close()
		}
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}

	if gw != nil {
		if err := gw.Close(); err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
			return fmt.Errorf("gzip close: %w", err)
		}
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

func marshalToWriter(w io.Writer, doc *Ableton, indent, includeXMLDecl bool) error {
	enc := xml.NewEncoder(w)
	if indent {
		enc.Indent("", "  ")
	}
	if err := enc.Encode(doc); err != nil {
		return fmt.Errorf("xml encode: %w", err)
	}
	if err := enc.Flush(); err != nil {
		return fmt.Errorf("xml flush: %w", err)
	}
	return nil
}
