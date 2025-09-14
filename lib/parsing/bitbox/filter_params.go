package bitbox

type FilterParams struct {
	Cutoff     int `xml:"cutoff,attr,omitempty"`
	Res        int `xml:"res,attr,omitempty"`
	FilterType int `xml:"filtertype,attr,omitempty"`
	FxTrigMode int `xml:"fxtrigmode,attr,omitempty"`
}
