package model

type TreeNode[T any] struct {
	Label      string        `json:"label,omitempty"`
	Selectable bool          `json:"selectable,omitempty"`
	Expandable bool          `json:"expandable,omitempty"`
	Children   []TreeNode[T] `json:"children,omitempty"`
	Data       T             `json:"data,omitempty"`
}
