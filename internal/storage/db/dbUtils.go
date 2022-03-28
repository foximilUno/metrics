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
	insertMetric           = `insert into metrics("name", "type", "value", "delta") values($1,$2,$3,$4);`
	getMetrics             = `select "name", "type", value, delta from metrics;`
	getMetricByNameAndType = `select "name", "type", value, delta from metrics where "name"=$1 and "type"=$2;`
	getMetricsByType       = `select "name", "type", value, delta from metrics where "type"=$1::varchar;`
	updateMetric           = `update metrics set value=$1, delta=$2 where "name"=$3;`
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

func GetAllMetricsFromDB(dbConn *sql.DB, request string, clause ...string) (map[string]*types.Metrics, error) {
	metrics := make(map[string]*types.Metrics)
	if dbConn == nil {
		return nil, fmt.Errorf("connetion to db is nil")
	}
	var rows *sql.Rows
	var err error
	//TODO да, ужасно))
	if len(clause) == 0 {
		rows, err = dbConn.Query(request)
	} else if len(clause) == 1 {
		rows, err = dbConn.Query(request, clause[0])
	} else {
		rows, err = dbConn.Query(request, clause[0], clause[1])
	}

	if err != nil {
		return nil, fmt.Errorf("cant exec getMetrics request: %e", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("cant clause rows: %e", err)
		}
	}(rows)

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

	return metrics, nil
}
