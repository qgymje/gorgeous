package gorgeous

import (
	"github.com/pkg/errors"
	"github.com/qgymje/gorgeous/provider"
)

type Option func(*Gorgeous) error

func WithLogger(l provider.ILogger) Option {
	return func(d *Gorgeous) error {
		if l == nil {
			return errors.New("logger is nil")
		}

		d.logger = l
		return nil
	}
}

func WithMetrics(name string, metrics provider.IMetrics) Option {
	return func(d *Gorgeous) error {
		if name == "" {
			return errors.New("metrics name is empty")
		}

		d.metricsName = name
		d.metrics = metrics
		return nil
	}
}
