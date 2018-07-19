package gorgeous

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fetcherDemo1 struct{}

func (f *fetcherDemo1) Name() string {
	return "fetcher demo"
}

func (f *fetcherDemo1) Size() int {
	return 2
}

func (f *fetcherDemo1) Action() (interface{}, error) {
	return "hello world", nil
}

func (f *fetcherDemo1) Interval() time.Duration {
	return 0
}

func (f *fetcherDemo1) Close() error {
	return nil
}

func Test_FetcherDemo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	f, err := NewFetcher(ctx, &fetcherDemo1{}, NewStdLogger(), NewStdMetrics())
	assert.Nil(t, err)
	f.Start()
	select {
	case data := <-f.Fetch():
		t.Logf("got data from fetcher: %v", data)
	}
	cancel()
}
