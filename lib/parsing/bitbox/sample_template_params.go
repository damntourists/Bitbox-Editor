package bitbox

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
