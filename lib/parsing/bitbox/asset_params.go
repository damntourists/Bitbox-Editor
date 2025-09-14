package bitbox

type AssetParams struct {
	RootNote       int `xml:"rootnote,attr,omitempty"`
	KeyRangeBottom int `xml:"keyrangebottom,attr,omitempty"`
	KeyRangeTop    int `xml:"keyrangetop,attr,omitempty"`
	AssSrcRow      int `xml:"asssrcrow,attr,omitempty"`
	AssSrcCol      int `xml:"asssrccol,attr,omitempty"`
}
