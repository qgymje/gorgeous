package provider

type IMetrics interface {
	Measure(tags map[string]string, fields map[string]interface{}) error
}
