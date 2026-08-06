package tests

import (
	"github.com/goravel/framework/testing"

	"github.com/nelsw/bytelyon/apps/mgr/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
