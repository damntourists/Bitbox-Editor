package bitbox

type DelayParams struct {
	DelayMuTime   int `xml:"delaymustime,attr,omitempty"`
	Feedback      int `xml:"feedback,attr,omitempty"`
	DelayBeatSync int `xml:"dealybeatsync,attr,omitempty"`
	Delay         int `xml:"delay,attr,omitempty"`
}
