package storage

import (
	"fmt"
	"github.com/foximilUno/metrics/internal/config"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/storage/db"
	"github.com/foximilUno/metrics/internal/storage/file"
	"github.com/foximilUno/metrics/internal/types"
	"log"
	"math"
	"time"
)

type MapStorage struct {
	Metrics map[string]*types.Metrics
	Persist repositories.Persist
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		Metrics: make(map[string]*types.Metrics),
	}
}

func (srm *MapStorage) Prepare(cfg *config.MetricServerConfig) error {
	var persist repositories.Persist
	var err error
	if len(cfg.DatabaseDsn) != 0 {
		persist, err = db.NewDbPersist(cfg.DatabaseDsn)
		if err != nil {
			return fmt.Errorf("problem with create db persist: %e", err)
		}
	} else if len(cfg.StoreFile) != 0 {
		persist = file.NewFilePersist(cfg.StoreFile)
	}

	if persist != nil {

		srm.Persist = persist

		if cfg.Restore {
			log.Printf("Restore from %s\r", persist)

			srm.Metrics, err = persist.Load()

			if err != nil {
				return fmt.Errorf("cant load metrics from persist: %e\n", err)
			}
		} else {
			log.Println("Start server without restoring from persist")
		}

		saveTicker := time.NewTicker(cfg.StoreInterval)

		go func(ticker *time.Ticker) {
			for {
				select {
				case <-ticker.C:
					if err := srm.Dump(); err != nil {
						log.Printf("cant save : err:%e", err)
					}
				default:
					time.Sleep(1 * time.Second)
				}
			}
		}(saveTicker)

	} else {
		log.Println("function \"Dump\" is turned off")
	}
	return nil
}

func (srm *MapStorage) Dump() error {
	return srm.Persist.Dump(srm.Metrics)
}

func (srm *MapStorage) SaveMetric(metric *types.Metrics) error {
	if metric.MType == "counter" {
		if _, ok := srm.Metrics[metric.ID]; !ok {
			srm.Metrics[metric.ID] = metric
			return nil
		}
		curDelta := srm.Metrics[metric.ID].Delta

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
		srm.Metrics[metric.ID] = metric

	} else {
		srm.Metrics[metric.ID] = metric
	}
	return nil
}

func (srm *MapStorage) GetGaugeMetricAsString(name string) (string, error) {
	if m, ok := srm.Metrics[name]; !ok {
		return "", fmt.Errorf("cant find such metric with type gauge")
	} else {
		if m.Value == nil {
			return "", nil
		}
		return fmt.Sprint(*m.Value), nil
	}
}

func (srm *MapStorage) GetCounterMetricAsString(name string) (string, error) {
	if m, ok := srm.Metrics[name]; !ok {
		return "", fmt.Errorf("cant find such metric with type counter")
	} else {
		if m.Delta == nil {
			return "", nil
		}
		return fmt.Sprint(*m.Delta), nil
	}
}

func (srm *MapStorage) GetMetricNamesByTypes(metricType string) []string {
	keys := make([]string, 0, len(srm.Metrics))
	for _, v := range srm.Metrics {
		if v.MType == metricType {
			keys = append(keys, v.ID)
		}
	}
	return keys
}

func (srm *MapStorage) IsPersisted() bool {
	return srm.Persist != nil
}

func sumWithCheck(var1 int64, var2 int64) (int64, error) {
	if math.MaxInt64-var1 >= var2 {
		return var1 + var2, nil
	} else {
		return 0, fmt.Errorf("overflow")
	}
}
