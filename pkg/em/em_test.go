package em

import (
	"testing"

	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	userID := ulid.MustParse("01KM01JC9PS1R4X4FDJNFAR4AZ")
	domain := "firefibers.com"
	e := &entity.Sitemap{
		Domain: domain,
		UserID: userID,
	}
	err := Find(e)
	assert.NoError(t, err)
	t.Log(e.SyncMap)
}
