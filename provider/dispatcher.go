package provider

type IDispatcher interface {
	Size() int
	Start()
	Stop()
}
