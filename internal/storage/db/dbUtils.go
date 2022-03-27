package db

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

	insertMetric = `insert into metrics("name", "type", "value", "delta") values($1,$2,$3,$4);`

	getMetrics = `select "name", "type", value, delta from metrics;`
)

func InsertMetricToDB(db *sql.DB, metrics *types.Metrics) (sql.Result, error) {
	return db.Exec(insertMetric, metrics.ID, metrics.MType, &metrics.Value, &metrics.Delta)
}

func GetDBConnectWithCreatedTable(connectionString string) (*sql.DB, error) {
	dbConnect, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}
	_, err = dbConnect.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	return dbConnect, nil
}

func GetAllMetricsFromDB(dbConn *sql.DB) (map[string]*types.Metrics, error) {
	metrics := make(map[string]*types.Metrics)
	if dbConn == nil {
		return nil, fmt.Errorf("connetion to db is nil")
	}
	rows, err := dbConn.Query(getMetrics)
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
