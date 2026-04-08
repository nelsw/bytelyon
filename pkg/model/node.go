package model

type Node struct {
	Label    string           `json:"label"`
	Children []*Node          `json:"children"`
	Nodes    map[string]*Node `json:"-"`
}
