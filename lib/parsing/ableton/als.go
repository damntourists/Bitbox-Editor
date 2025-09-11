package ableton

import "encoding/xml"

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
