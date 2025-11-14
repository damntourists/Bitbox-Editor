package bitbox

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
