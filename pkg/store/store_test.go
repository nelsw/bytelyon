package store

import (
	"fmt"
	"testing"

	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/shopify"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	logs.Init("trace")
	db, err := New[string, shopify.Order]("orders.json")
	assert.NoError(t, err)
	defer db.Close()

	fmt.Println(db)
}
