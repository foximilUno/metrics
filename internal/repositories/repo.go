package repositories

import "github.com/foximilUno/metrics/internal/types"

type MetricSaver interface {
	WithPersist(dbURL string) error
	Load() error
	Dump() error
	SaveMetric(metric *types.Metrics) error
	GetGaugeMetricAsString(name string) (string, error)
	GetCounterMetricAsString(name string) (string, error)
	GetMetricNamesByTypes(typeMetric string) []string
}
