package provider

// IWorker push data into worker
type IWorker interface {
	Start()
	Stop()
	Work() chan<- interface{}
	Next(IWorker)
}

type IWorkHandler interface {
	Name() string
	Size() int
	HandleData(interface{}) (interface{}, error)
	SetNext(IWorkHandler)
	Next() IWorkHandler
	Close() error
}
