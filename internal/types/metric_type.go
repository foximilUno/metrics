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
