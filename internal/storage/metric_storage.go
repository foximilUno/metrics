package storage

import (
	"database/sql"
	"fmt"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/types"
	_ "github.com/jackc/pgx/stdlib"
	"log"
	"math"
)

const (
	createTableQuery = `create table if not exists metrics(
	"id" int generated always as identity,
	"name" varchar(255),
	"type" varchar(50),
	"value" double precision,
	"delta" int
);`

	emptyMetricTable = `delete from metrics;`

	insertMetric = `insert into metrics("name", "type", "value", "delta") values($1,$2,$3,$4);`

	getMetrics = `select "name", "type", value, delta from metrics;`
)

type MapStorage struct {
	metrics map[string]*types.Metrics
	DB      *sql.DB
}

type Dump struct {
	DumpedMetrics []types.Metrics `json:"dumpedMetrics"`
}

func NewMapStorage() repositories.MetricSaver {

	return &MapStorage{
		DB:      nil,
		metrics: make(map[string]*types.Metrics),
	}
}

func (srm *MapStorage) WithPersist(dbURL string) error {
	dbConn, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	_, err = dbConn.Exec(createTableQuery)
	if err != nil {
		return err
	}

	srm.DB = dbConn
	return nil

}

func (srm *MapStorage) Load() error {

	rows, err := srm.DB.Query(getMetrics)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	if err != nil {
		return err
	}

	for rows.Next() {
		m := &types.Metrics{}
		var val sql.NullFloat64
		var delt sql.NullInt64
		err = rows.Scan(&m.ID, &m.MType, &val, &delt)
		if err != nil {
			return err
		}
		if val.Valid {
			m.Value = &val.Float64
		}
		if delt.Valid {
			m.Delta = &delt.Int64
		}

		srm.metrics[m.ID] = m
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	log.Printf("loaded %d metrics\n", len(srm.metrics))
	return nil
}

func (srm *MapStorage) Dump() error {
	log.Println("save to DB")

	_, err := srm.DB.Exec(emptyMetricTable)
	if err != nil {
		return err
	}
	for _, v := range srm.metrics {
		_, err = srm.DB.Exec(insertMetric, v.ID, v.MType, &v.Value, &v.Delta)
		if err != nil {
			return err
		}
	}

	return nil
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
