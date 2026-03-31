package model

import (
	"encoding/xml"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Time time.Time

func (v *Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		log.Warn().Bytes("b", b).Msg("failed to unmarshal time with format: " + time.RFC3339)
		t, err = time.Parse("2006-01-02T15:04:05Z", s)
	}
	if err != nil {
		log.Warn().Bytes("b", b).Msg("failed to unmarshal time with format: 2006-01-02T15:04:05Z")
		return err
	}
	*v = Time(t)
	return nil
}

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
func (v *Time) String() string          { return v.UTC().Format(time.RFC3339) }
func (v *Time) Before(t time.Time) bool { return v.UTC().Before(t.UTC()) }
func (v *Time) UTC() time.Time          { return time.Time(*v).UTC() }
