package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func ReadNewMetric(r *http.Request) (*Metrics, error) {
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
