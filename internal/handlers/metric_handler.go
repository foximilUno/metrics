package handlers

import (
	"fmt"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/go-chi/chi"
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

func SaveMetrics(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricVal := chi.URLParam(r, "metricValue")

		//check method only POST
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}
		////check elements in path
		//segments := strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")

		//if len(segments) != 4 {
		//	http.Error(w, "path must be pattern like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>", http.StatusBadRequest)
		//	return
		//}

		if len(metricName) == 0 {
			http.Error(w, "metric name cant be empty", http.StatusBadRequest)
		}

		switch metricType {
		case "gauge":
			val, err := strconv.ParseFloat(metricVal, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s некорректного типа: counter: float64", metricVal), http.StatusBadRequest)
				return
			}
			s.SaveGauge(metricName, val)
		case "counter":
			val, err := strconv.ParseInt(metricVal, 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s некорректного типа: counter: int64", metricVal), http.StatusBadRequest)
				return
			}
			s.SaveCounter(metricName, val)
		default:
			http.Error(w, fmt.Sprintf("Bad request: %s cant be, use %s", metricType, reflect.ValueOf(allowedTypes).MapKeys()), http.StatusNotImplemented)
			return
		}

		log.Printf("invoked update metric with type \"%s\" witn name \"%s\" with value \"%s\"", metricType, metricName, metricVal)
	}
}

func GetMetric(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if r.Method != http.MethodGet {
			http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
			return
		}

		if len(metricName) == 0 {
			http.Error(w, "metric name cant be empty", http.StatusBadRequest)
			return
		}

		var result string
		var err error
		switch metricType {
		case "gauge":
			result, err = s.GetGaugeMetricAsString(metricName)
		case "counter":
			result, err = s.GetCounterMetricAsString(metricName)
		default:
			http.Error(w, fmt.Sprintf("Bad request: %s cant be, use %s", metricType, reflect.ValueOf(allowedTypes).MapKeys()), http.StatusNotImplemented)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err = w.Write([]byte(result))
		if err != nil {
			log.Printf("error while write response: %e", err)
			return
		}
	}
}

func GetMetricsTable(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var response []byte
		response = append(response, []byte(preHTML)...)
		for _, v := range s.GetGaugeMetricNames() {
			n, err := s.GetGaugeMetricAsString(v)
			if err != nil {
				log.Printf("error while getting value: %e", err)
			} else {
				response = append(response, []byte(fmt.Sprintf(trPattern, "gauge", v, v, n))...)
			}
		}
		for _, v := range s.GetCounterMetricNames() {
			n, err := s.GetCounterMetricAsString(v)
			if err != nil {
				log.Printf("error while getting value: %e", err)
			} else {
				response = append(response, []byte(fmt.Sprintf(trPattern, "counter", v, v, n))...)
			}
		}
		response = append(response, []byte(postHTML)...)
		_, err := w.Write(response)
		if err != nil {
			return
		}
	}
}
