package ptr

var (
	True        = Of(true)
	False       = Of(false)
	ZeroFloat64 = Of(0.0)
)

func Of[T any](t T) *T {
	return &t
}
