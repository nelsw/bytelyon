package contract

type Worker interface {
	Start()
	Stop() error
}

type Job interface {
	Run(quit <-chan struct{}) error
}
