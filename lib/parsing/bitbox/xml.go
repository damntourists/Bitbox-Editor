package bitbox

import "encoding/xml"

type Document struct {
	XMLName xml.Name `xml:"document"`
	Session Session  `xml:"session"`
}

type Session struct {
	Cells []Cell `xml:"cell"`
}

type Cell struct {
	Row      *int   `xml:"row,attr,omitempty"`
	Column   *int   `xml:"column,attr,omitempty"`
	Layer    *int   `xml:"layer,attr,omitempty"`
	Filename string `xml:"filename,attr,omitempty"`
	Type     string `xml:"type,attr"`

	Params     ParamSet    `xml:"params"`
	ModSources []ModSource `xml:"modsource"`
	Slices     *Slices     `xml:"slices"`
}

type ParamSet map[string]string

type ModSource struct {
	Dest   string `xml:"dest,attr"`
	Src    string `xml:"src,attr"`
	Slot   *int   `xml:"slot,attr,omitempty"`
	Amount *int   `xml:"amount,attr,omitempty"`
}

type Slices struct{}
