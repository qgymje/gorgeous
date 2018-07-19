package gorgeous

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/qgymje/gorgeous/provider"
)

type task struct {
	name          string
	fetchHandlers []provider.IFetchHandler
	workHandlers  []provider.IWorkHandler
}

type Gorgeous struct {
	workerNumber int32
	stop         chan struct{}
	done         chan struct{}
	ctx          context.Context

	logger      provider.ILogger
	metricsName string
	metrics     provider.IMetrics

	tasks map[string]task
}

func New(ctx context.Context, opts ...Option) (*Gorgeous, error) {
	g := new(Gorgeous)
	g.ctx = ctx
	g.stop = make(chan struct{})
	g.done = make(chan struct{})
	g.tasks = make(map[string]task)

	for _, opt := range opts {
		if err := opt(g); err != nil {
			return nil, err
		}
	}

	if g.logger == nil {
		g.logger = NewStdLogger()
	}

	if g.metrics == nil {
		g.metrics = NewStdMetrics()
	}

	return g, nil
}

func (g *Gorgeous) Add(name string, fh provider.IFetchHandler, wh provider.IWorkHandler) {
	atomic.AddInt32(&g.workerNumber, 1)

	g.tasks[name] = task{
		name:          name,
		fetchHandlers: []provider.IFetchHandler{fh},
		workHandlers:  []provider.IWorkHandler{wh},
	}

}

func waitForTerminal(fn func()) {
	stopSignals := make(chan os.Signal, 1)
	signal.Notify(stopSignals, syscall.SIGINT, syscall.SIGTERM)
	<-stopSignals

	fn()
}

func (g *Gorgeous) Start(fn func()) error {
	for _, t := range g.tasks {
		if err := g.run(t); err != nil {
			return err
		}
	}

	waitForTerminal(func() {
		fn()
		g.shutdown()
	})
	return nil
}

func (g *Gorgeous) run(t task) error {
	fetchers := []provider.IFetcher{}
	workers := []provider.IWorker{}

	for _, fh := range t.fetchHandlers {
		f, err := NewFetcher(g.ctx, fh, g.logger, g.metrics)
		if err != nil {
			return err
		}
		fetchers = append(fetchers, f)
	}

	for _, wh := range t.workHandlers {
		w, err := NewWorker(g.ctx, wh, g.logger, g.metrics)
		if err != nil {
			return err
		}

		nextHandler := wh.Next()
		for nextHandler != nil {
			nw, err := NewWorker(g.ctx, nextHandler, g.logger, g.metrics)
			if err != nil {
				return err
			}
			nw.Start()
			w.Next(nw)
			w = nw
			nextHandler = nextHandler.Next()
		}

		workers = append(workers, w)
	}

	dispatch, err := NewDispatcher(
		g.ctx,
		t.name,
		fetchers,
		workers,
		g.logger,
		g.metrics,
	)
	if err != nil {
		return err
	}
	dispatch.Start()

	go func() {
		<-g.stop
		dispatch.Stop()
		g.done <- struct{}{}
	}()

	return nil
}

func (g *Gorgeous) shutdown() {
	go func() {
		var i int32
		l := atomic.LoadInt32(&g.workerNumber)
		for i = 0; i < l; i++ {
			g.stop <- struct{}{}
		}
	}()

	var q int32
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-g.done:
			q++
			if q == atomic.LoadInt32(&g.workerNumber) {
				return
			}
		case <-timeout:
			log.Println("stop timeout!")
			return
		}
	}
}
