package storage

import (
	"fmt"
	"github.com/foximilUno/metrics/internal/repositories"
	"log"
	"math"
)

type MapStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMapStorage() repositories.MetricSaver {
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

	after, err := sumWithCheck(srm.counters[name], val)
	if err != nil {
		log.Printf("cant increase counter with name %s: %e\r\n", name, err)
		return
	}
	log.Printf("successfully increase counter %s: before: %d, val:%d, after:%d \r\n", name, srm.counters[name], val, after)
	srm.counters[name] = after
}

func (srm *MapStorage) GetGaugeMetricAsString(name string) (string, error) {
	if val, ok := srm.gauges[name]; !ok {
		return "", fmt.Errorf("cant find such metric with type gauge")
	} else {
		return fmt.Sprint(val), nil
	}
}

func (srm *MapStorage) GetCounterMetricAsString(name string) (string, error) {
	if val, ok := srm.counters[name]; !ok {
		return "", fmt.Errorf("cant find such metric with type counter")
	} else {
		return fmt.Sprint(val), nil
	}
}

func (srm *MapStorage) GetGaugeMetricNames() []string {
	keys := make([]string, 0, len(srm.gauges))
	for k := range srm.gauges {
		keys = append(keys, k)
	}
	return keys
}

func (srm *MapStorage) GetCounterMetricNames() []string {
	keys := make([]string, 0, len(srm.counters))
	for k := range srm.counters {
		keys = append(keys, k)
	}
	return keys
}

func sumWithCheck(var1 int64, var2 int64) (int64, error) {
	if math.MaxInt64-var1 >= var2 {
		return var1 + var2, nil
	} else {
		return 0, fmt.Errorf("overflow")
	}
}
