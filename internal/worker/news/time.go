package news

import (
	"encoding/xml"
	"strings"
	"time"
)

type Time time.Time

func (v *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	s = strings.Trim(s, `"`) // Remove quotes from the JSON string
	if s == "" || s == "null" {
		return nil // Handle empty or null strings
	}

	t, err := time.Parse(time.RFC1123, s) // Parse using your custom format
	if err != nil {
		return err
	}

	*v = Time(t.UTC())
	return nil
}

func (v *Time) String() string {
	return v.UTC().Format(time.RFC3339)
}

func (v *Time) Before(t time.Time) bool {
	return time.Time(*v).UTC().Before(t)
}

func (v *Time) UTC() time.Time {
	return time.Time(*v).UTC()
}
