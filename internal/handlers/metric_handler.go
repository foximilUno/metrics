package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/foximilUno/metrics/internal/config"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/secure"
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

func readNewMetricByJSON(r *http.Request) (*types.Metrics, error) {
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

func readNewMetricByTextPlain(pathArray []string) (*types.Metrics, error) {
	metric := &types.Metrics{}

	if len(pathArray[2]) == 0 {
		return nil, fmt.Errorf("metric name cant be empty")
	}

	metric.MType = pathArray[1]
	metric.ID = pathArray[2]
	if len(pathArray) == 4 {
		switch metric.MType {
		case "gauge":
			val, err := strconv.ParseFloat(pathArray[3], 64)
			if err != nil {
				return nil, err
			}
			metric.Value = &val
		case "counter":
			val, err := strconv.ParseInt(pathArray[3], 10, 64)
			if err != nil {
				return nil, err
			}
			metric.Delta = &val
		}
	}
	return metric, nil
}

func returnData(w http.ResponseWriter, r *http.Request, data []byte, contentType string) error {
	var err error
	w.Header().Set("Content-Type", contentType)
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		var b bytes.Buffer
		gzC := gzip.NewWriter(&b)

		_, err = gzC.Write(data)
		if err != nil {
			return err
		}
		err = gzC.Close()
		if err != nil {
			return err
		}
		w.Header().Set("Content-Encoding", "gzip")
		_, err = w.Write(b.Bytes())
	} else {

		_, err = w.Write(data)
	}
	return err
}

func commonSaveMetric(metric *types.Metrics, s repositories.MetricSaver) (int, error) {
	switch metric.MType {
	case "gauge", "counter":
		if err := s.SaveMetric(metric); err != nil {
			return http.StatusInternalServerError, err
		}
	default:
		return http.StatusNotImplemented,
			fmt.Errorf("bad request: %s cant be, use %s",
				metric.ID,
				reflect.ValueOf(allowedTypes).MapKeys())
	}

	log.Printf("invoked update metric")
	if err := json.NewEncoder(log.Writer()).Encode(metric); err != nil {
		fmt.Println("cant encode metric object")
	}
	return 0, nil
}

// commonGetMetric return string value of metric and save metric value/delta to parameter object
func commonGetMetric(metric *types.Metrics, s repositories.MetricSaver) (string, int, error) {
	var result string
	var err error
	switch metric.MType {
	case "gauge":
		result, err = s.GetGaugeMetricAsString(metric.ID)
		if err != nil {
			return "", http.StatusNotFound, err
		}
		val, err := strconv.ParseFloat(result, 64)
		if err != nil {
			return "", http.StatusInternalServerError, err
		}
		metric.Value = &val
	case "counter":
		result, err = s.GetCounterMetricAsString(metric.ID)
		if err != nil {
			return "", http.StatusNotFound, err
		}
		val, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			return "", http.StatusInternalServerError, err
		}
		metric.Delta = &val
	default:
		return "",
			http.StatusNotImplemented,
			fmt.Errorf("bad request: %s cant be, use %s",
				metric.ID,
				reflect.ValueOf(allowedTypes).MapKeys())
	}
	return result, 0, nil
}

func isNeedEncryptOrDecrypt(cfg *config.MetricServerConfig) bool {
	return len(cfg.Key) > 0
}

func SaveMetricsViaTextPlain(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//check elements in path
		segments := strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")

		if len(segments) != 4 {
			http.Error(w, "path must be pattern like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>", http.StatusBadRequest)
			return
		}
		metric, err := readNewMetricByTextPlain(segments)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if status, err := commonSaveMetric(metric, s); err != nil {
			SendError(status, w, err.Error())
			return
		}
		w.WriteHeader(200)
	}
}

func SaveMetricsViaJSON(s repositories.MetricSaver, cfg *config.MetricServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := readNewMetricByJSON(r)
		if err != nil {
			SendError(http.StatusBadRequest, w, err.Error())
			return
		}

		if isNeedEncryptOrDecrypt(cfg) {
			isValid, err := secure.IsValidHash(metric.Format(), metric.Hash, cfg.Key)
			if err != nil {
				SendError(http.StatusBadRequest, w, err.Error())
				return
			}
			if !isValid {
				SendError(http.StatusBadRequest, w, "hash is not equal handled metric")
				return
			}
		}

		if status, err := commonSaveMetric(metric, s); err != nil {
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

		metric, err := readNewMetricByTextPlain(segments)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		strResult, httpStatus, err := commonGetMetric(metric, s)
		if err != nil {
			SendError(httpStatus, w, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(strResult))
		if err != nil {
			SendError(http.StatusInternalServerError, w, "error while writing")
		}
	}
}

func GetMetricViaJSON(s repositories.MetricSaver, cfg *config.MetricServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := readNewMetricByJSON(r)
		if err != nil {
			SendError(http.StatusBadRequest, w, err.Error())
			return
		}

		_, httpStatus, err := commonGetMetric(metric, s)
		if err != nil {
			SendError(httpStatus, w, err.Error())
			return
		}

		if isNeedEncryptOrDecrypt(cfg) {
			hash, err := secure.EncryptMetric(metric, cfg.Key)
			if err != nil {
				SendError(http.StatusInternalServerError, w, err.Error())
				return
			}
			metric.Hash = hash
		}

		bb, err := json.Marshal(metric)
		if err != nil {
			SendError(http.StatusInternalServerError, w, err.Error())
			return
		}
		err = returnData(w, r, bb, "application/json")
		if err != nil {
			SendError(http.StatusInternalServerError, w, "error while zipping")
			return
		}
	}
}

func GetMetricsTable(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var response []byte
		response = append(response, []byte(preHTML)...)
		for _, v := range s.GetMetricNamesByTypes("gauge") {
			n, _ := s.GetGaugeMetricAsString(v)
			response = append(response, []byte(fmt.Sprintf(trPattern, "gauge", v, v, n))...)
		}
		for _, v := range s.GetMetricNamesByTypes("counter") {
			n, _ := s.GetCounterMetricAsString(v)
			response = append(response, []byte(fmt.Sprintf(trPattern, "counter", v, v, n))...)
		}
		response = append(response, []byte(postHTML)...)
		err := returnData(w, r, response, "text/html")
		if err != nil {
			SendError(http.StatusInternalServerError, w, "error while zipping")
			return
		}
	}
}
