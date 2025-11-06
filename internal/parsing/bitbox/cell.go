package bitbox

import (
	"encoding/xml"
	"fmt"
	"io"
)

type Cell struct {
	Row      *int   `xml:"row,attr,omitempty"`
	Column   *int   `xml:"column,attr,omitempty"`
	Layer    *int   `xml:"layer,attr,omitempty"`
	Filename string `xml:"filename,attr,omitempty"`
	Name     string `xml:"name,attr,omitempty"`
	Type     string `xml:"type,attr"`
	Params   any    `xml:"-"`

	ModSources []ModSource   `xml:"modsource"`
	Slices     *Slices       `xml:"slices"`
	Sequence   *NoteSequence `xml:"-"` // TODO
}

func (c *Cell) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "row":
			var v int
			if _, err := fmt.Sscan(a.Value, &v); err == nil {
				c.Row = &v
			}
		case "column":
			var v int
			if _, err := fmt.Sscan(a.Value, &v); err == nil {
				c.Column = &v
			}
		case "layer":
			var v int
			if _, err := fmt.Sscan(a.Value, &v); err == nil {
				c.Layer = &v
			}
		case "filename":
			c.Filename = a.Value
		case "name":
			c.Name = a.Value
		case "type":
			c.Type = a.Value
		}
	}

	for {
		tok, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "params":
				paramsVal, err := newParamsForType(c.Type)
				if err != nil {
					return err
				}
				if err := d.DecodeElement(paramsVal, &t); err != nil {
					return err
				}
				c.Params = paramsVal

			case "modsource":
				var ms ModSource
				if err := d.DecodeElement(&ms, &t); err != nil {
					return err
				}
				c.ModSources = append(c.ModSources, ms)

			case "slices":
				var sl Slices
				if err := d.DecodeElement(&sl, &t); err != nil {
					return err
				}
				c.Slices = &sl

			case "sequence":
				var seq NoteSequence
				if err := d.DecodeElement(&seq, &t); err != nil {
					return err
				}
				c.Sequence = &seq

			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}

		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
	return nil
}

func (c *Cell) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "cell"
	start.Attr = start.Attr[:0]

	if c.Type != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "type"}, Value: c.Type})
	}
	if c.Row != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "row"}, Value: fmt.Sprint(*c.Row)})
	}
	if c.Column != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "column"}, Value: fmt.Sprint(*c.Column)})
	}
	if c.Layer != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "layer"}, Value: fmt.Sprint(*c.Layer)})
	}
	if c.Filename != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "filename"}, Value: c.Filename})
	}
	if c.Name != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "name"}, Value: c.Name})
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	if c.Params != nil {
		if err := e.EncodeElement(c.Params, xml.StartElement{Name: xml.Name{Local: "params"}}); err != nil {
			return err
		}
	}

	for _, ms := range c.ModSources {
		if err := e.EncodeElement(ms, xml.StartElement{Name: xml.Name{Local: "modsource"}}); err != nil {
			return err
		}
	}
	if c.Slices != nil {
		if err := e.EncodeElement(c.Slices, xml.StartElement{Name: xml.Name{Local: "slices"}}); err != nil {
			return err
		}
	}
	if c.Sequence != nil {
		if err := e.EncodeElement(c.Sequence, xml.StartElement{Name: xml.Name{Local: "sequence"}}); err != nil {
			return err
		}
	}

	if err := e.EncodeToken(start.End()); err != nil {
		return err
	}
	return e.Flush()
}
