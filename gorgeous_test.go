package gorgeous

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/qgymje/gorgeous/provider"
	"github.com/stretchr/testify/assert"
)

type fetcherDemo struct{}

func (f *fetcherDemo) Name() string {
	return "fetcher demo"
}

func (f *fetcherDemo) Size() int {
	return 1
}

func (f *fetcherDemo) Action() (interface{}, error) {
	return "hello world", nil
}

func (f *fetcherDemo) Interval() time.Duration {
	return 1 * time.Second
}

func (f *fetcherDemo) Close() error {
	log.Println("fetcher is closed.")
	return nil
}

type workerDemo struct{}

func (h *workerDemo) Name() string {
	return "demo"
}
func (h *workerDemo) Size() int {
	return 1
}

func (h *workerDemo) HandleData(data interface{}) (interface{}, error) {
	log.Printf("got data: %+v", data)
	return "data that pass to next worker", fmt.Errorf("some error: data = %+v", data)
}

func (h *workerDemo) SetNext(provider.IWorkHandler) {

}

func (h *workerDemo) Next() provider.IWorkHandler {
	return nil
}

func (h *workerDemo) Close() error {
	log.Println("worker is closed")
	return nil
}

func Test_startTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	gor, err := NewGorgeous(ctx)
	assert.Nil(t, err)

	gor.Add("demo", &fetcherDemo{}, &workerDemo{})

	var stop = func() {
		cancel()
		log.Printf("done")
	}

	if err := gor.Start(stop); err != nil {
		log.Fatal("gorgeous start failed:", err)
	}

}
