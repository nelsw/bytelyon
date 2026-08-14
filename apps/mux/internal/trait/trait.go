package trait

type HasRoute interface {
	Route() string
}

type HasScreenshot interface {
	Screenshot() (string, []byte, bool)
}

type HasContent interface {
	Content() (string, []byte, bool)
}

type HasWebRequest interface {
	Request() (string, map[string][]string, any)
}
