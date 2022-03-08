package storage

import (
	"encoding/json"
	"fmt"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/types"
	"io"
	"log"
	"math"
	"os"
)

type MapStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

type Dump struct {
	DumpedMetrics []types.Metrics `json:"dumpedMetrics"`
}

func NewMapStorage() repositories.MetricSaver {
	return &MapStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64)}
}

func (srm *MapStorage) LoadFromFile(filename string) error {

	fmt.Println("load", filename)

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	decoder := json.NewDecoder(file)

	var dump *Dump

	if err := decoder.Decode(&dump); err == io.EOF {
		return nil
	} else if err != nil {
		log.Fatal("fatal", err)
	}

	log.Printf("loaded %d metrics\n", len(dump.DumpedMetrics))

	for _, v := range dump.DumpedMetrics {
		switch v.MType {
		case "gauge":
			srm.SaveGauge(v.ID, *v.Value)
		case "counter":
			err := srm.SaveCounter(v.ID, *v.Delta)
			if err != nil {
				return err
			}
		default:
			log.Fatalf("not supported type of metric")
		}
	}
	return nil
}

func (srm *MapStorage) SaveToFile(filename string) error {
	log.Println("save to", filename)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)
	if err != nil {
		return err
	}

	metricsArray := &Dump{
		[]types.Metrics{},
	}

	for k, v := range srm.gauges {

		tempV := v
		metricsArray.DumpedMetrics = append(metricsArray.DumpedMetrics, types.Metrics{
			ID:    k,
			MType: "gauge",
			Value: &tempV,
		})
	}

	for k, v := range srm.counters {
		tempV := v
		metricsArray.DumpedMetrics = append(metricsArray.DumpedMetrics, types.Metrics{
			ID:    k,
			MType: "counter",
			Delta: &tempV,
		})
	}

	encoder := json.NewEncoder(file)
	err = encoder.Encode(metricsArray)
	if err != nil {
		return err
	}
	return nil
}

func (srm *MapStorage) SaveGauge(name string, val float64) {
	srm.gauges[name] = val
}

func (srm *MapStorage) SaveCounter(name string, val int64) error {
	if _, ok := srm.counters[name]; !ok {
		srm.counters[name] = 0
	}

	after, err := sumWithCheck(srm.counters[name], val)
	if err != nil {
		return fmt.Errorf("cant increase counter with name %s: %e", name, err)

	}
	log.Printf("successfully increase counter %s: before: %d, val:%d, after:%d \r\n", name, srm.counters[name], val, after)
	srm.counters[name] = after
	return nil
}

func (srm *MapStorage) GetGaugeMetricAsString(name string) (string, error) {
	if val, ok := srm.gauges[name]; !ok {
		return "", fmt.Errorf("cant find such metric with type gauge")
	} else {
		return fmt.Sprint(val), nil
	}
}

func (srm *MapStorage) GetCounterMetricAsString(name string) (string, error) {
	if val, ok := srm.counters[name]; !ok {
		return "", fmt.Errorf("cant find such metric with type counter")
	} else {
		return fmt.Sprint(val), nil
	}
}

func (srm *MapStorage) GetGaugeMetricNames() []string {
	keys := make([]string, 0, len(srm.gauges))
	for k := range srm.gauges {
		keys = append(keys, k)
	}
	return keys
}

func (srm *MapStorage) GetCounterMetricNames() []string {
	keys := make([]string, 0, len(srm.counters))
	for k := range srm.counters {
		keys = append(keys, k)
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
