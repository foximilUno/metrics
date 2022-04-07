package db

import (
	"database/sql"
	"fmt"
	"github.com/foximilUno/metrics/internal/types"
	"log"
)

const (
	emptyMetricTable = `delete from metrics;`
)

type dbPersist struct {
	DB *sql.DB
}

func NewDBPersist(connectionString string) (*dbPersist, error) {
	dbConnect, err := GetDBConnectWithCreatedTable(connectionString)
	if err != nil {
		return nil, err
	}
	return &dbPersist{
		DB: dbConnect,
	}, nil

}

func (dbs *dbPersist) Load() (map[string]*types.Metrics, error) {
	return GetAllMetricsFromDB(dbs.DB, getMetrics)
}

func (dbs *dbPersist) Dump(metrics map[string]*types.Metrics) error {
	log.Println("save to DB")

	if dbs.DB == nil {
		return fmt.Errorf("connetion to db is nil")
	}

	_, err := dbs.DB.Exec(emptyMetricTable)
	if err != nil {
		return err
	}
	for _, v := range metrics {
		_, err = InsertMetricToDB(dbs.DB, v)
		if err != nil {
			return err
		}
	}

	return nil
}
