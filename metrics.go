package gorgeous

import "log"

type StdMetrics struct{}

func NewStdMetrics() *StdMetrics {
	return &StdMetrics{}
}

func (m *StdMetrics) Measure(tags map[string]string, fields map[string]interface{}) error {
	log.Printf("[metrics]tags: %+v, fields: %+v]\n", tags, fields)
	return nil
}
