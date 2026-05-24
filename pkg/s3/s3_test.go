package s3

import (
	"path"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

func TestListDirectories(t *testing.T) {
	logs.Init("trace")
	ListDirectories("test")
}

func TestTheFuture(t *testing.T) {
	logs.Init("trace")
	DeleteDirectory("test")
	news()
	search()
	sitemap()
}

/*
> topic
>> urls (page)
*/
func news() {
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.NewsBotType, "iran", fakeUrl())+".json", []byte(`{}`))
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.NewsBotType, "iran", fakeUrl())+".json", []byte(`{}`))
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.NewsBotType, "iran", fakeUrl())+".json", []byte(`{}`))

	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.NewsBotType, "btc", fakeUrl())+".json", []byte(`{}`))
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.NewsBotType, "btc", fakeUrl())+".json", []byte(`{}`))

	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.NewsBotType, "eth", fakeUrl())+".json", []byte(`{}`))
}

/*
> query
>> id (group)
>>> urls (page)
*/
func search() {
	id := model.NewULID()
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.SearchBotType, "yoga mats", id, fakeUrl())+".json", []byte(`{}`))
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.SearchBotType, "yoga mats", id, fakeUrl())+".json", []byte(`{}`))

	id = model.NewULID()
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.SearchBotType, "yoga mats", id, fakeUrl())+".json", []byte(`{}`))
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.SearchBotType, "yoga mats", id, fakeUrl())+".json", []byte(`{}`))
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.SearchBotType, "yoga mats", id, fakeUrl())+".json", []byte(`{}`))

	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.SearchBotType, "bikes", id, fakeUrl())+".json", []byte(`{}`))

	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.SearchBotType, "books", id, fakeUrl())+".json", []byte(`{}`))
}

/*
> domain
>> urls (group)
>>> id (page)
*/
func sitemap() {

	url := fakeUrl()
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.SitemapBotType, "firefibers.com", url, model.NewULID())+".json", []byte(`{}`))
	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.SitemapBotType, "firefibers.com", url, model.NewULID())+".json", []byte(`{}`))

	PutPrivateObject(util.Path("test", "users", ulid.Zero, model.SitemapBotType, "bytelyon.com", fakeUrl())+".json", []byte(`{}`))
}

func fakeUrl() string {
	url := fakeDomain()
	if p := fakePath(); p != "" {
		url += "/" + p
	}
	if q := fakeQuery(); q != "" {
		url += "?" + q
	}
	return url
}

func fakeDomain() string {
	return faker.DomainName()
}

func fakePath() string {
	count := util.Between(0, 3)
	if count == 0 {
		return ""
	}

	var paths []string
	for i := 0; i < count; i++ {
		paths = append(paths, faker.Word())
	}

	return path.Join(paths...)
}

func fakeQuery() string {
	count := util.Between(0, 3)
	if count == 0 {
		return ""
	}

	var params []string
	for i := 0; i < count; i++ {
		param := faker.Word() + "=" + faker.Word()
		params = append(params, param)
	}
	return strings.Join(params, "&")
}
