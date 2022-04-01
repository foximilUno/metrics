package db

import (
	"database/sql"
	"fmt"
	"github.com/foximilUno/metrics/internal/storage/utils"
	"github.com/foximilUno/metrics/internal/types"
	"log"
)

type dbStorage struct {
	DB *sql.DB
}

func NewDBStorage(connectionString string) (*dbStorage, error) {
	dbConnect, err := GetDBConnectWithCreatedTable(connectionString)
	if err != nil {
		return nil, err
	}
	st := &dbStorage{
		dbConnect,
	}
	return st, nil
}

func (d *dbStorage) SaveMetric(metric *types.Metrics) error {

	//find metric with same ID and type
	metrics, err := GetAllMetricsFromDB(d.DB, getMetricByNameAndType, metric.ID, metric.MType)
	if err != nil {
		return err
	}
	//if metrics more than one is bad consistency
	if len(metrics) > 1 {
		return fmt.Errorf(
			"finded %d metrics with name=\"%s\" and type=\"%s\"", len(metrics), metric.ID, metric.MType)
	}

	//if no one same metric - is new - create new row in DB and exit from function
	if len(metrics) == 0 {
		_, err = InsertMetricToDB(d.DB, metric)
		if err != nil {
			return err
		}
		return nil
	}

	// by default change existed metric in map
	err = utils.ChangeMetricInMap(metrics, metric)
	if err != nil {
		return err
	}
	_, err = d.DB.Exec(updateMetric, &metrics[metric.ID].Value, &metrics[metric.ID].Delta, metric.ID, metric.MType)
	return err
}

//SaveBatchMetrics create DB write in one trancsaction
func (d *dbStorage) SaveBatchMetrics(metrics []*types.Metrics) error {
	//get all metrics from DB
	curDBMetrics, err := GetAllMetricsFromDB(d.DB, getMetrics)
	if err != nil {
		return err
	}

	notExistedMetrics := make([]*types.Metrics, 0, len(metrics))
	existedMetrics := make([]*types.Metrics, 0, len(metrics))
	for _, m := range metrics {
		//if no one metric
		if curM, ok := curDBMetrics[m.ID]; !ok {
			notExistedMetrics = append(notExistedMetrics, m)
			//if metric exist by type is not equal
		} else if m.MType != curM.MType {
			notExistedMetrics = append(notExistedMetrics, m)
			//if exist
		} else {
			existedMetrics = append(existedMetrics, m)
		}
	}

	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Printf("error while roolback: %e", err)
		}
	}(tx)
	insSt, err := tx.Prepare(insertMetric)
	if err != nil {
		return err
	}
	if len(notExistedMetrics) != 0 {
		for _, m := range metrics {
			_, err = insSt.Exec(m.ID, m.MType, &m.Value, &m.Delta)
			if err != nil {
				return err
			}
		}
	}
	updSt, err := tx.Prepare(updateMetric)
	if err != nil {
		return err
	}
	if len(existedMetrics) != 0 {
		for _, m := range existedMetrics {
			if err = utils.ChangeMetricInMap(curDBMetrics, m); err != nil {
				return err
			}
			_, err = updSt.Exec(&curDBMetrics[m.ID].Value, &curDBMetrics[m.ID].Delta, m.ID, m.MType)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (d *dbStorage) GetGaugeMetricAsString(name string) (string, error) {
	metrics, err := GetAllMetricsFromDB(d.DB, getMetricByNameAndType, name, "gauge")
	if err != nil {
		return "", err
	}
	return utils.GetMetricValueAsStringFromMap(metrics, name)
}

func (d *dbStorage) GetCounterMetricAsString(name string) (string, error) {
	metrics, err := GetAllMetricsFromDB(d.DB, getMetricByNameAndType, name, "counter")
	if err != nil {
		return "", err
	}
	return utils.GetMetricCounterAsStringFromMap(metrics, name)
}

func (d *dbStorage) GetMetricNamesByTypes(metricType string) []string {
	metrics, err := GetAllMetricsFromDB(d.DB, getMetricsByType, metricType)
	if err != nil {
		return []string{}
	}
	return utils.GetMetricNamesByTypesFromMap(metrics, metricType)
}

func (d *dbStorage) IsBatchSupports() bool {
	return true
}
