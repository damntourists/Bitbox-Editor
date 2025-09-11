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

// LiveSet - Holds tracks, master track and scene names.
type LiveSet struct {
	Tracks      Tracks      `xml:"Tracks"`
	MasterTrack MasterTrack `xml:"MasterTrack"`
	SceneNames  SceneNames  `xml:"SceneNames"`
}

// Tracks - AudioTracks
type Tracks struct {
	AudioTracks []AudioTrack `xml:"AudioTrack"`
}

// AudioTrack - information that defines an audio track
type AudioTrack struct {
	Id          int              `xml:"Id,attr"`
	Name        TrackName        `xml:"Name"`
	ColorIndex  AttrInt          `xml:"ColorIndex"`
	DeviceChain TrackDeviceChain `xml:"DeviceChain"`
}

// TrackName
type TrackName struct {
	EffectiveName AttrString `xml:"EffectiveName"`
	UserName      AttrString `xml:"UserName"`
}

// TrackDeviceChain - Device chains and sequencers
type TrackDeviceChain struct {
	MainSequencer   Sequencer         `xml:"MainSequencer"`
	FreezeSequencer Sequencer         `xml:"FreezeSequencer"`
	DeviceChain     *InnerDeviceChain `xml:"DeviceChain"`
}

// InnerDeviceChain - Device chains and sequencers
type InnerDeviceChain struct {
	Devices *struct{} `xml:"Devices"`
}

// Sequencer clip slots
type Sequencer struct {
	ClipSlotList ClipSlotList `xml:"ClipSlotList"`
}

// ClipSlotList - Clip slots
type ClipSlotList struct {
	ClipSlots []ClipSlot `xml:"ClipSlot"`
}

// The XML has nested <ClipSlot><HasStop/><ClipSlot><Value>...</Value></ClipSlot></ClipSlot>
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

// AudioClip configuration
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

// WarpMarkers - WarpMarker collection
type WarpMarkers struct {
	WarpMarker []WarpMarker `xml:"WarpMarker"`
}

// WarpMarker -
type WarpMarker struct {
	SecTime  float64 `xml:"SecTime,attr"`
	BeatTime float64 `xml:"BeatTime,attr"`
}

// SampleRef - contains a file reference
type SampleRef struct {
	FileRef FileRef `xml:"FileRef"`
}

// FileRef - File reference
type FileRef struct {
	HasRelativePath  AttrBool     `xml:"HasRelativePath"`
	RelativePathType AttrInt      `xml:"RelativePathType"`
	RelativePath     RelativePath `xml:"RelativePath"`
	Name             AttrString   `xml:"Name"`
}

// RelativePath - Joinable relative path elements
type RelativePath struct {
	Elements []RelativePathElement `xml:"RelativePathElement"`
}

type RelativePathElement struct {
	Dir string `xml:"Dir,attr"`
}

// ClipLoop - Defines the loop points and loop state of the clip.
type ClipLoop struct {
	LoopStart       AttrFloat64 `xml:"LoopStart"`
	LoopEnd         AttrFloat64 `xml:"LoopEnd"`
	StartRelative   AttrFloat64 `xml:"StartRelative"`
	LoopOn          AttrBool    `xml:"LoopOn"`
	OutMarker       AttrFloat64 `xml:"OutMarker"`
	HiddenLoopStart AttrFloat64 `xml:"HiddenLoopStart"`
	HiddenLoopEnd   AttrFloat64 `xml:"HiddenLoopEnd"`
}

// ClipTimeSignature - Time signature
type ClipTimeSignature struct {
	TimeSignatures TimeSignatures `xml:"TimeSignatures"`
}

// TimeSignatures - List of remotable time signatures
type TimeSignatures struct {
	Remoteable []RemoteableTimeSignature `xml:"RemoteableTimeSignature"`
}

// RemoteableTimeSignature - Time signature
type RemoteableTimeSignature struct {
	Numerator   AttrInt     `xml:"Numerator"`
	Denominator AttrInt     `xml:"Denominator"`
	Time        AttrFloat64 `xml:"Time"`
}

// MasterTrack - tempo, time signature automation
type MasterTrack struct {
	DeviceChain MasterDeviceChain `xml:"DeviceChain"`
}

// MasterDeviceChain holds the master mixer.
type MasterDeviceChain struct {
	Mixer MasterMixer `xml:"Mixer"`
}

// MasterMixer - tempo, time signature automation
type MasterMixer struct {
	Tempo         TempoAutomation         `xml:"Tempo"`
	TimeSignature TimeSignatureAutomation `xml:"TimeSignature"`
}

// TempoAutomation - tempo automation
type TempoAutomation struct {
	ArrangerAutomation ArrangerAutomationFloat `xml:"ArrangerAutomation"`
}

// TimeSignatureAutomation - time signature automation
type TimeSignatureAutomation struct {
	ArrangerAutomation ArrangerAutomationEnum `xml:"ArrangerAutomation"`
}

// ArrangerAutomationFloat - Holds float events
type ArrangerAutomationFloat struct {
	Events FloatEvents `xml:"Events"`
}

// ArrangerAutomationEnum - Holds enum events
type ArrangerAutomationEnum struct {
	Events EnumEvents `xml:"Events"`
}

// FloatEvents - List of float event
type FloatEvents struct {
	FloatEvent []FloatEvent `xml:"FloatEvent"`
}

// EnumEvents - List of enum event
type EnumEvents struct {
	EnumEvent []EnumEvent `xml:"EnumEvent"`
}

// FloatEvent -	Float event information
type FloatEvent struct {
	Time  float64 `xml:"Time,attr"`
	Value float64 `xml:"Value,attr"`
}

// EnumEvent -	Enum event information
type EnumEvent struct {
	Time  float64 `xml:"Time,attr"`
	Value int     `xml:"Value,attr"`
}

// SceneNames - Holds a list of scenes
type SceneNames struct {
	Scenes []Scene `xml:"Scene"`
}

type Scene struct{}

// Attribute wrapper types ------------------------------------------------------

// Many elements in this schema are like <X Value="..."/>
// These helper types keep your main structs concise.

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
