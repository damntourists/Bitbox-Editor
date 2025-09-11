package bitbox

import "encoding/xml"

type Document struct {
	XMLName xml.Name `xml:"document"`
	Session Session  `xml:"session"`
}

// Session holds all <cell> elements.
type Session struct {
	Cells []Cell `xml:"cell"`
}

// Cell is a polymorphic node. Depending on Type, it may have different params.
type Cell struct {
	// Attributes (some are optional depending on the cell)
	Row      *int   `xml:"row,attr,omitempty"`
	Column   *int   `xml:"column,attr,omitempty"`
	Layer    *int   `xml:"layer,attr,omitempty"`
	Filename string `xml:"filename,attr,omitempty"`
	Type     string `xml:"type,attr"`

	// Children
	Params     ParamSet    `xml:"params"`
	ModSources []ModSource `xml:"modsource"`
	Slices     *Slices     `xml:"slices"`
}

// ParamSet captures all attributes on <params .../> into a map.
// Example: p["gaindb"] == "0", p["recthresh"] == "-20000", etc.
type ParamSet map[string]string

// ModSource corresponds to <modsource dest="..." src="..." slot="..." amount="..."/>
type ModSource struct {
	Dest   string `xml:"dest,attr"`
	Src    string `xml:"src,attr"`
	Slot   *int   `xml:"slot,attr,omitempty"`
	Amount *int   `xml:"amount,attr,omitempty"`
}

// Slices corresponds to an empty <slices/> element.
type Slices struct{}
