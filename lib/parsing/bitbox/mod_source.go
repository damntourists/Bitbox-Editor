package bitbox

type ModSource struct {
	Dest   string `xml:"dest,attr"`
	Src    string `xml:"src,attr"`
	Slot   *int   `xml:"slot,attr,omitempty"`
	Amount *int   `xml:"amount,attr,omitempty"`
}
