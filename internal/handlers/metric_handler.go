package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/foximilUno/metrics/internal/repositories"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
)

const (
	//defaultApplicationType = "text/plain"
	preHTML   = `<html><header></header><body><div><table border="solid"><caption>Metrics</caption><tr><th>metricName</th><th>metricVal</th></tr>`
	postHTML  = `</table></div></body>`
	trPattern = `<tr><td><a href="/value/%s/%s">%s</a></td><td>%s</td></tr>`
)

var allowedTypes = map[string]string{
	"gauge":   "gauge",
	"counter": "counter",
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func readNewMetric(r *http.Request) (*Metrics, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read request body: %e", err)
	}
	var metric *Metrics
	err = json.Unmarshal(bodyBytes, &metric)

	if err != nil {
		return nil, fmt.Errorf("can't read request body: %e", err)
	}
	return metric, nil
}

func SaveMetrics(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//check method only POST
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		metric, err := readNewMetric(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if metric.Delta == nil && metric.Value == nil {
			http.Error(w, "delta or value must not be empty", http.StatusBadRequest)
			return
		}

		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				http.Error(w, "value cant be empty", http.StatusBadRequest)
				return
			}
			s.SaveGauge(metric.ID, *metric.Value)
		case "counter":
			if metric.Delta == nil {
				http.Error(w, "delta cant be empty", http.StatusBadRequest)
				return
			}
			err = s.SaveCounter(metric.ID, *metric.Delta)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, fmt.Sprintf("Bad request: %s cant be, use %s", metric.ID, reflect.ValueOf(allowedTypes).MapKeys()), http.StatusNotImplemented)
			return
		}

		log.Printf("invoked update metric with type \"%s\" witn name \"%s\" with value \"%d\"|\"%d\"", metric.MType, metric.ID, metric.Value, metric.Delta)

		w.WriteHeader(200)
	}
}

func GetMetric(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		metric, err := readNewMetric(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var result string
		switch metric.MType {
		case "gauge":
			result, err = s.GetGaugeMetricAsString(metric.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			val, err := strconv.ParseFloat(result, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			metric.Value = &val
		case "counter":
			result, err = s.GetCounterMetricAsString(metric.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			val, err := strconv.ParseInt(result, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			metric.Delta = &val
		default:
			http.Error(w, fmt.Sprintf("Bad request: %s cant be, use %s", metric.MType, reflect.ValueOf(allowedTypes).MapKeys()), http.StatusNotImplemented)
			return
		}

		bb, err := json.Marshal(metric)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, err = w.Write(bb)
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetMetricsTable(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var response []byte
		response = append(response, []byte(preHTML)...)
		for _, v := range s.GetGaugeMetricNames() {
			n, _ := s.GetGaugeMetricAsString(v)
			response = append(response, []byte(fmt.Sprintf(trPattern, "gauge", v, v, n))...)
		}
		for _, v := range s.GetCounterMetricNames() {
			n, _ := s.GetCounterMetricAsString(v)
			response = append(response, []byte(fmt.Sprintf(trPattern, "counter", v, v, n))...)
		}
		response = append(response, []byte(postHTML)...)
		_, err := w.Write(response)
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
