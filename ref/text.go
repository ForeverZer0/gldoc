package ref

import (
	"bytes"
	"encoding/xml"

	"github.com/ForeverZer0/gldoc/util"
)

// A Text object captures the inner XML as text, but also flattens nested elements into their inner text value.
type Text string

// UnmarshalXML implements the [xml.Unmarshaler] interface.
func (text *Text) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var sb bytes.Buffer
	token, err := d.Token()

	for token != start.End() && err == nil {
		if chardata, ok := token.(xml.CharData); ok {
			sb.Write(chardata)
		}
		token, err = d.Token()
	}
	*text = Text(util.Sanitize(sb.String()))
	return err
}
