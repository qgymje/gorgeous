package gorgeous

import (
	"context"
	"time"

	"github.com/qgymje/gorgeous/provider"
)

var _ provider.IDispatcher = (*Dispatcher)(nil)

// Dispatcher receive data from fetcher
// collect related objects to do the job
type Dispatcher struct {
	ctx  context.Context
	name string
	size int

	fetchers []provider.IFetcher
	workers  []provider.IWorker

	metrics provider.IMetrics
	logger  provider.ILogger

	data chan interface{}
	done chan struct{}
	err  chan error
}

// NewDispatcher create a new workerpool
func NewDispatcher(ctx context.Context, name string, fetchers []provider.IFetcher, workers []provider.IWorker, logger provider.ILogger, metrics provider.IMetrics) (*Dispatcher, error) {
	d := new(Dispatcher)
	d.ctx = ctx
	d.name = name
	d.fetchers = fetchers
	d.workers = workers
	d.logger = logger
	d.metrics = metrics

	var size int
	for _, f := range fetchers {
		size += f.Size()
	}

	d.size = size
	d.data = make(chan interface{})
	d.done = make(chan struct{}, d.size)
	d.err = make(chan error, 1)
	return d, nil
}

func (d *Dispatcher) Size() int {
	return d.size
}

// SetDebug set dispatch debug model
// Start start to dispatch worker
func (d *Dispatcher) Start() {
	for _, f := range d.fetchers {
		f.Start()
	}

	for _, w := range d.workers {
		w.Start()
	}

	d.mergedInput()

	for i := 0; i < d.size; i++ {
		supervisor(worker(d.run), 2000)
	}
}

// Stop goroutines
func (d *Dispatcher) Stop() {
	for _, f := range d.fetchers {
		f.Stop()
	}

	counter := 0
	for range d.done {
		counter++
		if counter == d.size {
			break
		}
	}
	close(d.data)
	close(d.err)
	close(d.done)

	for _, w := range d.workers {
		w.Stop()
	}
}

func (d *Dispatcher) mergedInput() {
	for _, f := range d.fetchers {
		go func(f provider.IFetcher) {
			for msg := range f.Fetch() {
				d.data <- msg
			}
		}(f)
	}
}

func (d *Dispatcher) run() {
	d.logger.Debugf("dispatcher: %s is running.", d.name)

	for {
		select {
		case <-d.ctx.Done():
			d.done <- struct{}{}
			return

		case v := <-d.data:
			start := time.Now()

			for _, w := range d.workers {
				w.Work() <- v
			}
			d.measure(start)

		case err := <-d.err:
			d.handleError(err)
		}
	}
}

func (d *Dispatcher) handleError(err error) {
	d.logger.Errorf("dispatcher: %s got an error:%s.", d.name, err)
}

func (d *Dispatcher) measure(start time.Time) {
	duration := time.Since(start).Nanoseconds() / 1e3 // us

	d.metrics.Measure(
		map[string]string{
			"dispatcher": d.name,
		},
		map[string]interface{}{
			"duration": duration,
		},
	)
}
