package repositories

type MetricSaver interface {
	SaveGauge(name string, val float64)
	SaveCounter(name string, val int64)
}
