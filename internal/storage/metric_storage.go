package storage

import (
	"fmt"
	"github.com/foximilUno/metrics/internal/config"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/storage/file"
	"github.com/foximilUno/metrics/internal/storage/utils"
	"github.com/foximilUno/metrics/internal/types"
	"log"
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
	if len(cfg.StoreFile) != 0 {
		persist = file.NewFilePersist(cfg.StoreFile)
	}

	if persist != nil {

		srm.Persist = persist

		if cfg.Restore {
			log.Printf("Restore from %s\r", persist)

			srm.Metrics, err = persist.Load()

			if err != nil {
				return fmt.Errorf("cant load metrics from persist: %e", err)
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
	return utils.ChangeMetricInMap(srm.Metrics, metric)
}

func (srm *MapStorage) GetGaugeMetricAsString(name string) (string, error) {
	return utils.GetMetricValueAsStringFromMap(srm.Metrics, name)
}

func (srm *MapStorage) GetCounterMetricAsString(name string) (string, error) {
	return utils.GetMetricCounterAsStringFromMap(srm.Metrics, name)
}

func (srm *MapStorage) GetMetricNamesByTypes(metricType string) []string {
	return utils.GetMetricNamesByTypesFromMap(srm.Metrics, metricType)
}

func (srm *MapStorage) SaveBatchMetrics(metrics []*types.Metrics) error {
	return fmt.Errorf("not implemented for map storage")
}

func (d *MapStorage) IsBatchSupports() bool {
	return false
}
