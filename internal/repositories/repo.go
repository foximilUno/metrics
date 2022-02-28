package repositories

type MetricSaver interface {
	SaveGauge(name string, val float64)
	SaveCounter(name string, val int64) error
	GetGaugeMetricAsString(name string) (string, error)
	GetCounterMetricAsString(name string) (string, error)
	GetGaugeMetricNames() []string
	GetCounterMetricNames() []string
}
