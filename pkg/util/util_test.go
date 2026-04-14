package util

import (
	"errors"
	"testing"
)

func Test_Supress_String(t *testing.T) {
	v := Suppress[string]("", errors.New("err"))
	t.Logf("v=[%s]", v)
}
