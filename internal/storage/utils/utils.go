package utils

import (
	"fmt"
	"github.com/foximilUno/metrics/internal/types"
	"log"
	"math"
)

func GetMetricValueAsStringFromMap(metrics map[string]*types.Metrics, name string) (string, error) {
	if m, ok := metrics[name]; !ok {
		return "", fmt.Errorf("cant find such metric with type gauge")
	} else {
		if m.Value == nil {
			return "", nil
		}
		return fmt.Sprint(*m.Value), nil
	}
}

func GetMetricCounterAsStringFromMap(metrics map[string]*types.Metrics, name string) (string, error) {
	if m, ok := metrics[name]; !ok {
		return "", fmt.Errorf("cant find such metric with type counter")
	} else {
		if m.Delta == nil {
			return "", nil
		}
		return fmt.Sprint(*m.Delta), nil
	}
}

func GetMetricNamesByTypesFromMap(metrics map[string]*types.Metrics, typeMetric string) []string {
	keys := make([]string, 0, len(metrics))
	for _, v := range metrics {
		if v.MType == typeMetric {
			keys = append(keys, v.ID)
		}
	}
	return keys
}

func ChangeMetricInMap(metrics map[string]*types.Metrics, metric *types.Metrics) error {
	if metric.MType == "counter" {
		if _, ok := metrics[metric.ID]; !ok {
			metrics[metric.ID] = metric
			return nil
		}
		curDelta := metrics[metric.ID].Delta

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
		metrics[metric.ID] = metric

	} else {
		metrics[metric.ID] = metric
	}
	return nil
}

func sumWithCheck(var1 int64, var2 int64) (int64, error) {
	if math.MaxInt64-var1 >= var2 {
		return var1 + var2, nil
	} else {
		return 0, fmt.Errorf("overflow")
	}
}
