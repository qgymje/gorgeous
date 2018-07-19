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

type workerDemo1 struct {
	next provider.IWorkHandler
}

func (h *workerDemo1) Name() string {
	return "demo"
}
func (h *workerDemo1) Size() int {
	return 2
}

func (h *workerDemo1) HandleData(data interface{}) (interface{}, error) {
	log.Printf("got data: %+v", data)
	return "data that pass to next worker", fmt.Errorf("some error: data = %+v", data)
}

func (h *workerDemo1) Next() provider.IWorkHandler {
	return h.next
}

func (h *workerDemo1) Close() error {
	log.Println("demo is closed")
	return nil
}

type workerDemo2 struct {
}

func (h *workerDemo2) Name() string {
	return "demo2"
}
func (h *workerDemo2) Size() int {
	return 2
}

func (h *workerDemo2) HandleData(data interface{}) (interface{}, error) {
	log.Printf("got data: %+v", data)
	return nil, fmt.Errorf("some error: data = %+v", data)
}

func (h *workerDemo2) Next() provider.IWorkHandler {
	return nil
}

func (h *workerDemo2) Close() error {
	log.Println("demo2 is closed")
	return nil
}

func TestSpawnWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	w, err := NewWorker(ctx, &workerDemo1{}, NewStdLogger(), nil)
	assert.Nil(t, err)
	w.Start()
	w.Work() <- "some data"
	time.Sleep(1e9)
	cancel()
}

func TestSpawnWorkerWithNextWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wh1 := &workerDemo1{}
	wh2 := &workerDemo2{}
	wh1.next = wh2

	w, err := NewWorker(ctx, wh1, NewStdLogger(), NewStdMetrics())
	assert.Nil(t, err)
	w2, err := NewWorker(ctx, wh2, NewStdLogger(), nil)

	w.Start()
	w2.Start()
	w.Work() <- "some data"
	time.Sleep(1e9)
	cancel()
}
