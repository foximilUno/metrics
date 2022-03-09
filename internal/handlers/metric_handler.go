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

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func readNewMetric(r *http.Request) (*Metrics, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("can't read request body: %e", err)
	}
	var metric *Metrics
	err = json.Unmarshal(bodyBytes, &metric)

	if err != nil {
		return nil, fmt.Errorf("can't unmarshall request body: %e", err)
	}
	return metric, nil
}

func (m *Metrics) UnmarshalJSON(bytes []byte) error {
	var requestOnj map[string]interface{}
	err := json.Unmarshal(bytes, &requestOnj)
	if err != nil {
		return err
	}
	if v, ok := requestOnj["id"]; ok {
		m.ID = v.(string)
	}
	if v, ok := requestOnj["type"]; ok {
		m.MType = v.(string)
	}
	if v, ok := requestOnj["delta"]; ok {
		tempDelta := int64(v.(float64))
		m.Delta = &tempDelta
	}
	if v, ok := requestOnj["value"]; ok {
		tempValue := v.(float64)
		m.Value = &tempValue
	}
	return nil
}

type ResultError struct {
	Error string `json:"error"`
}

//func SendErrorWithString(w http.ResponseWriter, stringVal string) error {
//	err := json.NewEncoder(w).Encode(
//		&ResultError{
//			stringVal,
//		})
//	if err != nil {
//		return err
//	}
//	return nil
//}

//func SendErrorWithError(w http.ResponseWriter, errorVal error) error {
//	err := json.NewEncoder(w).Encode(
//		&ResultError{
//			errorVal.Error(),
//		})
//	if err != nil {
//		return err
//	}
//	return nil
//}

func SaveMetricsViaTextPlain(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//check method only POST
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		////check content type only defaultApplicationType
		//if r.Header.Get("Content-type") != defaultApplicationType {
		//	w.Header().Add("Allowed", "text/plain")
		//	http.Error(w, "Allowed text/plain only", http.StatusUnsupportedMediaType)
		//	return
		//}

		//check elements in path
		segments := strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")

		if len(segments) != 4 {
			http.Error(w, "path must be pattern like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>", http.StatusBadRequest)
			return
		}

		if len(segments[2]) == 0 {
			http.Error(w, "metric name cant be empty", http.StatusBadRequest)
		}

		switch segments[1] {
		case "gauge":
			val, err := strconv.ParseFloat(segments[3], 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s некорректного типа: counter: float64", segments[2]), http.StatusBadRequest)
				return
			}
			s.SaveGauge(segments[2], val)
		case "counter":
			val, err := strconv.ParseInt(segments[3], 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s некорректного типа: counter: int64", segments[2]), http.StatusBadRequest)
				return
			}
			err = s.SaveCounter(segments[2], val)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, fmt.Sprintf("Bad request: %s cant be, use %s", segments[1], reflect.ValueOf(allowedTypes).MapKeys()), http.StatusNotImplemented)
			return
		}

		log.Printf("invoked update metric with type \"%s\" witn name \"%s\" with value \"%s\"", segments[1], segments[2], segments[3])

		w.WriteHeader(200)
	}
}

func SaveMetricsViaJSON(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			//err := SendErrorWithString(w, "only POST allowed")
			//if err != nil {
			//	log.Println(err)
			//}
			return
		}

		metric, err := readNewMetric(r)

		if err != nil {
			log.Println("error reaDMetric:", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			//err := SendErrorWithError(w, err)
			//if err != nil {
			//	log.Println(err)
			//}
			return
		}

		if metric.Delta == nil && metric.Value == nil {
			w.WriteHeader(http.StatusBadRequest)

			//err := SendErrorWithString(w, "delta or value must not be empty")
			//if err != nil {
			//	log.Println(err)
			//}
			log.Printf("error metric.Delta == nil && metric.Value == nil: %e\n", err)

			return
		}

		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				//err := SendErrorWithString(w, "value cant be empty")
				//if err != nil {
				//	log.Println(err)
				//}
				log.Printf("error metric.Value == nil%e\n", err)
				return
			}
			s.SaveGauge(metric.ID, *metric.Value)
		case "counter":
			if metric.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				//err := SendErrorWithString(w, "delta cant be empty")
				//if err != nil {
				//	log.Println(err)
				//}
				log.Printf("error metric.Delta == nil %e\n", err)
				return
			}
			err = s.SaveCounter(metric.ID, *metric.Delta)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				//err := SendErrorWithError(w, err)
				//if err != nil {
				//	log.Println(err)
				//}

				log.Printf("error SaveCounter %e\n", err)
				return
			}
		default:
			w.WriteHeader(http.StatusNotImplemented)

			//err := SendErrorWithString(w, fmt.Sprintf("bad request: %s cant be, use %s", metric.ID, reflect.ValueOf(allowedTypes).MapKeys()))
			//if err != nil {
			//	log.Println(err)
			//}
			log.Printf("error StatusNotImplemented %e\n", err)
			return
		}

		log.Printf("invoked update metric with type \"%s\" witn name \"%s\" with value \"%d\"|\"%d\"", metric.MType, metric.ID, metric.Value, metric.Delta)

		w.WriteHeader(200)
	}
}

func GetMetricViaTextPlain(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		//check elements in path
		segments := strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")

		if len(segments) != 3 {
			http.Error(w, "path must be pattern like /value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>", http.StatusBadRequest)
			return
		}

		if len(segments[2]) == 0 {
			http.Error(w, "metric name cant be empty", http.StatusBadRequest)
			return
		}

		var result string
		var err error
		switch segments[1] {
		case "gauge":
			result, err = s.GetGaugeMetricAsString(segments[2])
		case "counter":
			result, err = s.GetCounterMetricAsString(segments[2])
		default:
			http.Error(w, fmt.Sprintf("Bad request: %s cant be, use %s", segments[1], reflect.ValueOf(allowedTypes).MapKeys()), http.StatusNotImplemented)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		_, err = w.Write([]byte(result))
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetMetricViaJSON(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)

			//err := SendErrorWithString(w, "only POST allowed")
			//if err != nil {
			//	log.Println(err)
			//}
			return
		}

		metric, err := readNewMetric(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			//
			//err := SendErrorWithError(w, err)
			//if err != nil {
			//	log.Println(err)
			//}
			return
		}

		var result string
		switch metric.MType {
		case "gauge":
			result, err = s.GetGaugeMetricAsString(metric.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)

				//err := SendErrorWithError(w, err)
				//if err != nil {
				//	log.Println(err)
				//}
				return
			}
			val, err := strconv.ParseFloat(result, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				//err := SendErrorWithError(w, err)
				//if err != nil {
				//	log.Println(err)
				//}
				return
			}
			metric.Value = &val
		case "counter":
			result, err = s.GetCounterMetricAsString(metric.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				//err := SendErrorWithError(w, err)
				//if err != nil {
				//	log.Println(err)
				//}
				return
			}
			val, err := strconv.ParseInt(result, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				//err := SendErrorWithError(w, err)
				//if err != nil {
				//	log.Println(err)
				//}
				return
			}
			metric.Delta = &val
		default:
			w.WriteHeader(http.StatusNotImplemented)

			//err := SendErrorWithString(w, fmt.Sprintf("bad request: %s cant be, use %s", metric.MType, reflect.ValueOf(allowedTypes).MapKeys()))
			//if err != nil {
			//	log.Println(err)
			//}
			return
		}

		bb, err := json.Marshal(metric)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			//err := SendErrorWithError(w, err)
			//if err != nil {
			//	log.Println(err)
			//}
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
