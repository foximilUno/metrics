package storage

type MapStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		make(map[string]float64),
		make(map[string]int64)}
}

func (srm *MapStorage) SaveGauge(name string, val float64) {
	srm.gauges[name] = val
}

func (srm *MapStorage) SaveCounter(name string, val int64) {
	if _, ok := srm.counters[name]; !ok {
		srm.counters[name] = 0
	}
	srm.counters[name] += val
}
