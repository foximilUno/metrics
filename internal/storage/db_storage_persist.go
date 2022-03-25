package storage

import (
	"database/sql"
	"fmt"
	"github.com/foximilUno/metrics/internal/types"
	"log"
)

const (
	createTableQuery = `create table if not exists metrics(
	"id" int generated always as identity,
	"name" varchar(255),
	"type" varchar(50),
	"value" double precision,
	"delta" bigint
);`

	emptyMetricTable = `delete from metrics;`

	insertMetric = `insert into metrics("name", "type", "value", "delta") values($1,$2,$3,$4);`

	getMetrics = `select "name", "type", value, delta from metrics;`
)

type dbStorage struct {
	DB *sql.DB
}

func NewDbPersist(connectionString string) (*dbStorage, error) {
	dbConnect, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}
	_, err = dbConnect.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	return &dbStorage{
		DB: dbConnect,
	}, nil

}

func (dbs *dbStorage) Load() (map[string]*types.Metrics, error) {
	metrics := make(map[string]*types.Metrics)
	if dbs.DB == nil {
		return nil, fmt.Errorf("connetion to db is nil")
	}
	rows, err := dbs.DB.Query(getMetrics)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	if err != nil {
		return nil, fmt.Errorf("cant exec getMetrics request: %e", err)
	}

	for rows.Next() {
		m := &types.Metrics{}
		var val sql.NullFloat64
		var delt sql.NullInt64
		err = rows.Scan(&m.ID, &m.MType, &val, &delt)
		if err != nil {
			return nil, fmt.Errorf("cant scan data: %e", err)
		}
		if val.Valid {
			m.Value = &val.Float64
		}
		if delt.Valid {
			m.Delta = &delt.Int64
		}

		metrics[m.ID] = m
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("there is errors while fetching: %e", err)
	}

	log.Printf("loaded %d metrics\n", len(metrics))
	return metrics, nil
}

func (dbs *dbStorage) Dump(metrics map[string]*types.Metrics) error {
	log.Println("save to DB")

	if dbs.DB == nil {
		return fmt.Errorf("connetion to db is nil")
	}

	_, err := dbs.DB.Exec(emptyMetricTable)
	if err != nil {
		return err
	}
	for _, v := range metrics {
		_, err = dbs.DB.Exec(insertMetric, v.ID, v.MType, &v.Value, &v.Delta)
		if err != nil {
			return err
		}
	}

	return nil
}
