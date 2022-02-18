package handlers

import (
	"fmt"
	"github.com/foximilUno/metrics/internal/repositories"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const (
	defaultApplicationType = "text/plain"
	servicePath            = "update"
)

var allowedTypes = map[string]string{
	"gauge":   "gauge",
	"counter": "counter",
}

func SaveMetrics(s repositories.MetricSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//check method only POST
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		//check content type only defaultApplicationType
		if r.Header.Get("Content-type") != defaultApplicationType {
			w.Header().Add("Allowed", "text/plain")
			http.Error(w, "Allowed text/plain only", http.StatusUnsupportedMediaType)
			return
		}

		//check elements in path
		segments := strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")
		if len(segments) != 4 {
			http.Error(w, "path must be pattern like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>", http.StatusBadRequest)
			return
		}
		//check first path segment
		if segments[0] != servicePath {
			http.Error(w, "path must be pattern like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>", http.StatusBadRequest)
			return
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
			s.SaveCounter(segments[2], val)
		default:
			http.Error(w, fmt.Sprintf("Bad request: %s cant be, use %s", segments[1], reflect.ValueOf(allowedTypes).MapKeys()), http.StatusBadRequest)
			return
		}

		log.Printf("invoked update metric with type \"%s\" witn name \"%s\" with value \"%s\"", segments[1], segments[2], segments[3])

		w.WriteHeader(200)
	}
}
