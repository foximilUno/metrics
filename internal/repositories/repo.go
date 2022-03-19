package repositories

type MetricSaver interface {
	Load(filename string) error
	Dump(filename string) error
	SaveGauge(name string, val float64)
	SaveCounter(name string, val int64) error
	GetGaugeMetricAsString(name string) (string, error)
	GetCounterMetricAsString(name string) (string, error)
	GetGaugeMetricNames() []string
	GetCounterMetricNames() []string
}
