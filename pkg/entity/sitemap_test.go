package entity

import (
	"fmt"
	"path"
	"testing"
)

func TestSitemap_Find(t *testing.T) {

	p := "https://firefibers.com/blog/2022/01/01/fire-fibers-is-now-open-source/foo.pdf"
	fmt.Println(path.Base(p))
	fmt.Println(path.Clean(p))
	fmt.Println(path.Dir(p))
	fmt.Println(path.Ext(p))
	fmt.Println(path.Split(p))
}
