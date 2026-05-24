package util

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
)

func Test_Supress_String(t *testing.T) {
	u, err := url.Parse("FireFibers.com/wat")
	if err != nil {
		panic(err)
	}
	fmt.Println(u.String())
	b, _ := json.MarshalIndent(u, "", "\t")
	fmt.Println(string(b))
}

func TestOr(t *testing.T) {
	var a any
	a = nil
	fmt.Println(Or(a))
}
