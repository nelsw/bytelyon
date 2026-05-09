package model

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

type Time time.Time

func (v *Time) IsZero() bool            { return v.UTC().IsZero() }
func (v *Time) String() string          { return v.UTC().Format(time.RFC3339) }
func (v *Time) Before(t time.Time) bool { return v.UTC().Before(t.UTC()) }
func (v *Time) ULID() ulid.ULID         { return NewULID(v.UTC()) }
func (v *Time) UTC() time.Time          { return time.Time(*v).UTC() }

func (v *Time) ToAttributeValue() types.AttributeValue {
	return &types.AttributeValueMemberS{Value: v.String()}
}

func (v *Time) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", v.String())), nil
}

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
	*v = Time(t.UTC())
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

func Now() *Time {
	return util.Ptr(Time(time.Now().UTC()))
}

func ParseTime(a any) (*Time, error) {
	log.Trace().
		Any("a", a).
		Type("t", a).
		Msg("parsing time")
	if a == nil {
		return nil, errors.New("cannot parse time; given: nil")
	}

	switch a.(type) {
	case time.Time:
		return util.Ptr(Time(a.(time.Time).UTC())), nil
	case *types.AttributeValueMemberS:
		return ParseTime(a.(*types.AttributeValueMemberS).Value)
	case string:
		if t, err := time.Parse(time.RFC3339, a.(string)); err != nil {
			return nil, err
		} else {
			return ParseTime(t)
		}
	}

	return nil, fmt.Errorf("cannot parse time; given: %v", a)
}
