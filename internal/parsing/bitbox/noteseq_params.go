package bitbox

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
