package contracts

type Entity[T any] interface {
	Key() string
}
