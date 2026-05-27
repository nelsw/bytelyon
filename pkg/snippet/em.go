package snippet

import (
	"github.com/nelsw/bytelyon/pkg/page"
)

func (m *Model) Save() (err error) {
	if err = page.SaveObject(m.URL, m.ID, m); err != nil {
		return
	} else if err = page.SaveScreenshot(m.URL, m.ID, m.Screenshot); err != nil {
		return
	}
	return
}
