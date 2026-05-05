package model

import "github.com/oklog/ulid/v2"

type Profile struct {
	// ID is the unique identifier for the profile
	ID ulid.ULID `json:"id"`
	// Name is the name of the user
	Name string `json:"name"`
}
