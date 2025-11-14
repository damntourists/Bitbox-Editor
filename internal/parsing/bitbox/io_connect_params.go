package bitbox

type IOConnectInParams struct {
	InputIOCon string `xml:"inputiocon,attr,omitempty"`
}

type IOConnectOutParams struct {
	OutputIOCon string `xml:"outputiocon,attr,omitempty"`
}
