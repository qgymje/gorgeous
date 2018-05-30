package fetcher

import (
	"context"
	"testing"
	"time"

	"git.verystar.cn/GaomingQian/gorgeous"
	"github.com/stretchr/testify/assert"
)

type fetcherDemo struct{}

func (f *fetcherDemo) Name() string {
	return "fetcher demo"
}

func (f *fetcherDemo) Size() int {
	return 2
}

func (f *fetcherDemo) Action() (interface{}, error) {
	return "hello world", nil
}

func (f *fetcherDemo) Interval() time.Duration {
	return 0
}

func (f *fetcherDemo) Close() error {
	return nil
}

func Test_FetcherDemo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	f, err := NewFetcher(ctx, &fetcherDemo{}, WithLogger(gorgeous.NewStdLogger()), WithMetrics(gorgeous.NewStdMetrics()))
	assert.Nil(t, err)
	f.Start()
	select {
	case data := <-f.Fetch():
		t.Logf("got data from fetcher: %v", data)
	}
	cancel()
}
