package model

type Node struct {
	Depth    int     `json:"-"`
	Label    string  `json:"label"`
	URL      string  `json:"url,omitempty"`
	Children []*Node `json:"children,omitempty"`
}
