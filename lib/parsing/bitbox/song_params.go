package bitbox

type SongParams struct {
	GlobTempo int `xml:"globtempo,attr,omitempty"`
	SongMode  int `xml:"songmode,attr,omitempty"`
	SectCount int `xml:"sectcount,attr,omitempty"`
	SectLoop  int `xml:"sectloop,attr,omitempty"`
	Swing     int `xml:"swing,attr,omitempty"`
	KeyMode   int `xml:"keymode,attr,omitempty"`
	KeyRoot   int `xml:"keyroot,attr,omitempty"`
}
