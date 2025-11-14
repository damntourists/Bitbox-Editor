package bitbox

type ReverbParams struct {
	Decay    int `xml:"decay,attr,omitempty"`
	PreDelay int `xml:"predelay,attr,omitempty"`
	Damping  int `xml:"damping,attr,omitempty"`
}
