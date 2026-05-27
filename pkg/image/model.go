package image

type Models []Model

type Model map[string]string

func Make(args ...string) Model {
	m := Model{
		"src":     "",
		"altText": "",
	}
	if len(args) > 0 {
		m.SetSrc(args[0])
	}
	if len(args) > 1 {
		m.SetAlt(args[1])
	}
	return m
}

func (m Model) GetSrc() string  { return m["src"] }
func (m Model) GetAlt() string  { return m["altText"] }
func (m Model) SetSrc(s string) { m["src"] = s }
func (m Model) SetAlt(s string) { m["altText"] = s }
