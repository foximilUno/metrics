package db

import (
	"database/sql"
	"fmt"
	"github.com/foximilUno/metrics/internal/storage/utils"
	"github.com/foximilUno/metrics/internal/types"
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

func (d dbStorage) SaveMetric(metric *types.Metrics) error {

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

func (d dbStorage) GetGaugeMetricAsString(name string) (string, error) {
	metrics, err := GetAllMetricsFromDB(d.DB, getMetricByNameAndType, name, "gauge")
	if err != nil {
		return "", err
	}
	return utils.GetMetricValueAsStringFromMap(metrics, name)
}

func (d dbStorage) GetCounterMetricAsString(name string) (string, error) {
	metrics, err := GetAllMetricsFromDB(d.DB, getMetricByNameAndType, name, "counter")
	if err != nil {
		return "", err
	}
	return utils.GetMetricCounterAsStringFromMap(metrics, name)
}

func (d dbStorage) GetMetricNamesByTypes(metricType string) []string {
	metrics, err := GetAllMetricsFromDB(d.DB, getMetricsByType, metricType)
	if err != nil {
		return []string{}
	}
	return utils.GetMetricNamesByTypesFromMap(metrics, metricType)
}
