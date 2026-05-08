package main

import (
	"fmt"
	"strings"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/model"
)

func main() {
	logs.Init("debug")

	m, err := db.Get(&model.Sitemap{Domain: "FireFibers.com"})
	if err != nil {
		panic(err)
	}

	var u string
	for _, url := range m.URLs.Slice() {
		// 2011-12-06T17:38:21+00:00
		u += fmt.Sprintf(`
	<url>
		<loc>%s</loc>
		<lastmod>%s</lastmod>
		<priority>0.5</priority>
		<changefreq>weekly</changefreq>
	</url>`, strings.ToLower(url), m.UpdatedAt)

	}

	a := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`

	z := `
</urlset>`

	fmt.Println(a + u + z)
	//quit := make(chan os.Signal, 1)
	//
	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	//log.Info().Msg("listening for quit signal (Ctrl+C)")
	//<-quit
	//fmt.Println()
	//
	//log.Info().Msg("quitting")
	//log.Info().Msg("exiting")
}
