package bitbox

type WavFile struct {
	Original string
	Resolved string

	Row, Column, Layer *int
	Type               string
}
