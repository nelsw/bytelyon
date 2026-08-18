package model

type Pageable interface {
	Route
	Screenshot() (string, []byte, bool)
}

type Route interface {
	Path() string
}

type Screenshot interface {
	Object
}

type Content interface {
	Object
}

type Object interface {
	Input() (string, []byte, bool)
}
