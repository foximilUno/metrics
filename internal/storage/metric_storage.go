package storage

import (
	"fmt"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/types"
	"log"
	"math"
)

type MapStorage struct {
	metrics map[string]*types.Metrics
	persist repositories.Persist
}

type Dump struct {
	DumpedMetrics []types.Metrics `json:"dumpedMetrics"`
}

func NewMapStorage() repositories.MetricSaver {
	return &MapStorage{
		metrics: make(map[string]*types.Metrics),
	}
}

func (srm *MapStorage) WithPersist(persist repositories.Persist) error {
	srm.persist = persist
	return nil
}

func (srm *MapStorage) Load() error {
	a, err := srm.persist.Load()
	if err != nil {
		return err
	}
	srm.metrics = a
	return nil
}

func (srm *MapStorage) Dump() error {
	return srm.persist.Dump(srm.metrics)
}

func (srm *MapStorage) SaveMetric(metric *types.Metrics) error {
	if metric.MType == "counter" {
		if _, ok := srm.metrics[metric.ID]; !ok {
			srm.metrics[metric.ID] = metric
			return nil
		}
		curDelta := srm.metrics[metric.ID].Delta

		after, err := sumWithCheck(*curDelta, *metric.Delta)
		if err != nil {
			return fmt.Errorf("cant increase counter with name %s: %e", metric.ID, err)

		}
		log.Printf(
			"successfully increase counter %s: before: %d, val:%d, after:%d \r\n",
			metric.ID,
			*curDelta,
			*metric.Delta,
			after)
		metric.Delta = &after
		srm.metrics[metric.ID] = metric

	} else {
		srm.metrics[metric.ID] = metric
	}
	return nil
}

func (srm *MapStorage) GetGaugeMetricAsString(name string) (string, error) {
	if m, ok := srm.metrics[name]; !ok {
		return "", fmt.Errorf("cant find such metric with type gauge")
	} else {
		if m.Value == nil {
			return "", nil
		}
		return fmt.Sprint(*m.Value), nil
	}
}

func (srm *MapStorage) GetCounterMetricAsString(name string) (string, error) {
	if m, ok := srm.metrics[name]; !ok {
		return "", fmt.Errorf("cant find such metric with type counter")
	} else {
		if m.Delta == nil {
			return "", nil
		}
		return fmt.Sprint(*m.Delta), nil
	}
}

func (srm *MapStorage) GetMetricNamesByTypes(metricType string) []string {
	keys := make([]string, 0, len(srm.metrics))
	for _, v := range srm.metrics {
		if v.MType == metricType {
			keys = append(keys, v.ID)
		}
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
