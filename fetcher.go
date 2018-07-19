package gorgeous

import (
	"context"
	"time"

	"github.com/qgymje/gorgeous/provider"
)

type Fetcher struct {
	ctx  context.Context
	name string
	size int

	logger  provider.ILogger
	metrics provider.IMetrics
	handler provider.IFetchHandler

	data chan interface{}
	err  chan error
	done chan struct{}
}

func NewFetcher(ctx context.Context, handler provider.IFetchHandler, logger provider.ILogger, metrics provider.IMetrics) (*Fetcher, error) {
	f := new(Fetcher)
	f.ctx = ctx
	f.name = handler.Name()
	f.size = handler.Size()
	f.handler = handler
	f.logger = logger
	f.metrics = metrics

	f.data = make(chan interface{})
	f.err = make(chan error, 1)
	f.done = make(chan struct{}, f.size)

	return f, nil
}

func (f *Fetcher) Fetch() <-chan interface{} {
	return f.data
}

func (f *Fetcher) Size() int {
	return f.size
}

func (f *Fetcher) Start() {
	for i := 0; i < f.size; i++ {
		supervisor(worker(f.run), 2000)
	}
}

func (f *Fetcher) Stop() {
	counter := 0
	for range f.done {
		counter++
		if counter == f.size {
			break
		}
	}

	close(f.data)
	close(f.err)
	close(f.done)

	f.handler.Close()

	f.logger.Debugf("fetcher: %s is done.", f.name)
}

func (f *Fetcher) run() {
	f.logger.Debugf("fetcher: %s is running.", f.name)

	if f.handler.Interval() == 0 {

		inputData := make(chan interface{})
		go func() {
			for {
				if data, err := f.handler.Action(); err != nil {
					f.err <- err
				} else {
					if data != nil {
						inputData <- data
					}
				}
				start := time.Now()
				f.measure(start)
			}
		}()

		for {
			select {
			case <-f.ctx.Done():
				f.done <- struct{}{}
				return

			case err := <-f.err:
				f.logger.Errorf("%s got an error: %+v", f.name, err)

			case data := <-inputData:
				f.data <- data
			}
		}
	} else {
		tick := time.NewTicker(f.handler.Interval())

		for {
			select {
			case <-f.ctx.Done():
				f.done <- struct{}{}
				return

			case err := <-f.err:
				f.logger.Errorf("%s got an error: %+v", f.name, err)

			case <-tick.C:
				start := time.Now()
				if data, err := f.handler.Action(); err != nil {
					f.err <- err
				} else {
					if data != nil {
						f.data <- data
					}
				}
				f.measure(start)
			}
		}
	}
}

func (f *Fetcher) measure(start time.Time) {
	duration := time.Since(start).Nanoseconds() / 1e3 // us

	f.metrics.Measure(
		map[string]string{
			"fetcher": f.name,
		},
		map[string]interface{}{
			"duration": duration,
		},
	)
}
