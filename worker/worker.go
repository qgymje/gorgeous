package worker

import (
	"context"
	"time"

	"git.verystar.cn/GaomingQian/gorgeous/provider"
	"git.verystar.cn/GaomingQian/gorgeous/supervisor"
)

type Option func(w *Worker) error

func WithLogger(l provider.ILogger) Option {
	return func(w *Worker) error {
		w.logger = l
		return nil
	}
}

func WithMetrics(metrics provider.IMetrics) Option {
	return func(w *Worker) error {
		w.metrics = metrics
		return nil
	}
}

type Worker struct {
	ctx  context.Context
	name string
	size int

	metrics provider.IMetrics
	logger  provider.ILogger

	handler provider.IWorkHandler

	hasNext    bool
	nextWorker provider.IWorker

	data chan interface{}
	err  chan error
	done chan struct{}
}

func NewWorker(ctx context.Context, handler provider.IWorkHandler, opts ...Option) (*Worker, error) {
	w := new(Worker)
	w.ctx = ctx
	w.handler = handler
	w.name = handler.Name()
	w.size = handler.Size()
	w.data = make(chan interface{})
	w.err = make(chan error, 1)
	w.done = make(chan struct{}, w.size)

	for _, opt := range opts {
		if err := opt(w); err != nil {
			return nil, err
		}
	}

	return w, nil
}

func (w *Worker) Work() chan<- interface{} {
	return w.data
}

func (w *Worker) Start() {
	for i := 0; i < w.size; i++ {
		supervisor.Supervisor(supervisor.Worker(w.run), 2000)
	}
}

func (w *Worker) Stop() {
	count := 0
	for range w.done {
		count++
		if count == w.size {
			break
		}
	}
	close(w.data)
	close(w.err)
	close(w.done)

	w.handler.Close()

	if w.hasNext {
		w.nextWorker.Stop()
	}
	w.logger.Debugf("worker: %s is done.", w.name)
}

func (w *Worker) Next(wk provider.IWorker) {
	w.nextWorker = wk
	w.hasNext = true
}

func (w *Worker) run() {
	w.logger.Debugf("worker: %s is running.", w.name)

	for {
		select {
		case <-w.ctx.Done():
			w.done <- struct{}{}
			return

		case data := <-w.data:
			start := time.Now()
			nextData, err := w.handler.HandleData(data)
			if err != nil {
				w.err <- err
			}
			if w.hasNext {
				w.nextWorker.Work() <- nextData
			}
			w.measure(start)

		case err := <-w.err:
			w.logger.Errorf("worker: %s,%s,%+v", time.Now().Format(time.RFC3339), w.name, err)
		}
	}
}

func (w *Worker) measure(start time.Time) {
	duration := time.Since(start).Nanoseconds() / 1e3 // us

	w.metrics.Measure(
		map[string]string{
			"worker": w.name,
		},
		map[string]interface{}{
			"duration": duration,
		},
	)
}
