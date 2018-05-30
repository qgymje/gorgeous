package fetcher

import (
	"context"
	"errors"
	"time"

	"git.verystar.cn/GaomingQian/gorgeous/provider"
	"git.verystar.cn/GaomingQian/gorgeous/supervisor"
)

type Option func(f *Fetcher) error

func WithLogger(l provider.ILogger) Option {
	return func(f *Fetcher) error {
		if l == nil {
			return errors.New("logger is nil")
		}

		f.logger = l
		return nil
	}
}

func WithMetrics(metrics provider.IMetrics) Option {
	return func(f *Fetcher) error {
		f.metrics = metrics
		return nil
	}
}

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

func NewFetcher(ctx context.Context, handler provider.IFetchHandler, opts ...Option) (*Fetcher, error) {
	f := new(Fetcher)
	f.ctx = ctx
	f.name = handler.Name()
	f.size = handler.Size()
	f.handler = handler

	f.data = make(chan interface{})
	f.err = make(chan error, 1)
	f.done = make(chan struct{}, f.size)

	for _, opt := range opts {
		if err := opt(f); err != nil {
			return nil, err
		}
	}

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
		supervisor.Supervisor(supervisor.Worker(f.run), 2000)
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
