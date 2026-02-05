package model

import "time"

type Article struct {
	URL       string
	Title     string
	Source    string
	Published time.Time
}
