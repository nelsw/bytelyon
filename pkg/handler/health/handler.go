package health

import "github.com/nelsw/bytelyon/pkg/api"

func Handler(r api.Request) (api.Response, error) { return r.NC(), nil }
