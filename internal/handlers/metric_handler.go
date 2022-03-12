package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/types"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const (
	preHTML   = `<html><header></header><body><div><table border="solid"><caption>Metrics</caption><tr><th>metricName</th><th>metricVal</th></tr>`
	postHTML  = `</table></div></body>`
	trPattern = `<tr><td><a href="/value/%s/%s">%s</a></td><td>%s</td></tr>`
)

var allowedTypes = map[string]string{
	"gauge":   "gauge",
	"counter": "counter",
}

type ResultError struct {
	Error string
}

func SendError(httpStatusCode int, w http.ResponseWriter, stringVal string) {
	w.WriteHeader(httpStatusCode)
	err := json.NewEncoder(w).Encode(
		&ResultError{
			stringVal,
		})
	if err != nil {
		log.Println(err)
	}
}

func ReadNewMetricByJSON(r *http.Request) (*types.Metrics, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read request body: %e", err)
	}
	defer r.Body.Close()
	var metric *types.Metrics
	err = json.Unmarshal(bodyBytes, &metric)

	if err != nil {
		return nil, fmt.Errorf("can't unmarshall request body: %e", err)
	}
	return metric, nil
}

func ReadNewMetricByTextPlain(pathArray []string) (*types.Metrics, error) {
	metric := &types.Metrics{}

	if len(pathArray[2]) == 0 {
		return nil, errors.New("metric name cant be empty")
	}

	metric.MType = pathArray[1]
	metric.ID = pathArray[2]
	switch metric.MType {
	case "gauge":
		val, err := strconv.ParseFloat(pathArray[3], 64)
		if err != nil {
			log.Println(err)
		}
		metric.Value = &val
	case "counter":
		val, err := strconv.ParseInt(pathArray[3], 10, 64)
		if err != nil {
			log.Println(err)
		}
		metric.Delta = &val
	}

	return metric, nil
}

func CommonSaveMetric(metric *types.Metrics, s repositories.MetricSaver) (error, int) {
	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			return errors.New("value cant be empty"), http.StatusBadRequest
		}
		s.SaveGauge(metric.ID, *metric.Value)
	case "counter":
		if metric.Delta == nil {
			return errors.New("delta cant be empty"), http.StatusBadRequest
		}
		if err := s.SaveCounter(metric.ID, *metric.Delta); err != nil {
			return err, http.StatusNotImplemented
		}
	default:
		return errors.New(fmt.Sprintf(
				"bad request: %s cant be, use %s",
				metric.ID,
				reflect.ValueOf(allowedTypes).MapKeys())),
			http.StatusBadRequest
	}

	log.Printf("invoked update metric")
	if err := json.NewEncoder(log.Writer()).Encode(metric); err != nil {
		fmt.Println("cant encode metric object")
	}
	return nil, 0
}

func CommonGetMetric(metric *types.Metrics, s repositories.MetricSaver) (string, error, int) {
	var result string
	var err error
	switch metric.MType {
	case "gauge":
		result, err = s.GetGaugeMetricAsString(metric.ID)
		if err != nil {
			return "", err, http.StatusNotFound
		}
		val, err := strconv.ParseFloat(result, 64)
		if err != nil {
			return "", err, http.StatusInternalServerError
		}
		metric.Value = &val
	case "counter":
		result, err = s.GetCounterMetricAsString(metric.ID)
		if err != nil {
			return "", err, http.StatusNotFound
		}
		val, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			return "", err, http.StatusInternalServerError
		}
		metric.Delta = &val
	default:
		return "",
			errors.New(fmt.Sprintf(
				"Bad request: %s cant be, use %s",
				metric.ID,
				reflect.ValueOf(allowedTypes).MapKeys())),
			http.StatusNotImplemented
	}
	return result, nil, 0
}

func SaveMetricsViaTextPlain(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//check elements in path
		segments := strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")

		if len(segments) != 4 {
			http.Error(w, "path must be pattern like /value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>", http.StatusBadRequest)
			return
		}
		metric, err := ReadNewMetricByTextPlain(segments)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err, status := CommonSaveMetric(metric, s); err != nil {
			SendError(status, w, err.Error())
		}
		w.WriteHeader(200)
	}
}

func SaveMetricsViaJSON(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := ReadNewMetricByJSON(r)
		if err != nil {
			SendError(http.StatusBadRequest, w, err.Error())
			return
		}

		if err, status := CommonSaveMetric(metric, s); err != nil {
			SendError(status, w, err.Error())
		}
		w.WriteHeader(200)
	}
}

func GetMetricViaTextPlain(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//check elements in path
		segments := strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")
		if len(segments) != 3 {
			http.Error(w, "path must be pattern like /value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>", http.StatusBadRequest)
			return
		}

		metric, err := ReadNewMetricByTextPlain(segments)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		strResult, err, httpStatus := CommonGetMetric(metric, s)
		if err != nil {
			SendError(httpStatus, w, err.Error())
		}

		_, err = w.Write([]byte(strResult))
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetMetricViaJSON(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := ReadNewMetricByJSON(r)
		if err != nil {
			SendError(http.StatusBadRequest, w, err.Error())
			return
		}

		_, err, httpStatus := CommonGetMetric(metric, s)
		if err != nil {
			SendError(httpStatus, w, err.Error())
		}

		bb, err := json.Marshal(metric)
		if err != nil {
			SendError(http.StatusInternalServerError, w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
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
