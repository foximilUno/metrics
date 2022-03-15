package types

import "log"

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func (m *Metrics) Equal(otherMetric *Metrics) bool {
	var compers bool
	switch m.MType {
	case "gauge":
		if m.Value == nil && otherMetric.Value == nil {
			compers = true
		} else if m.Value != nil && otherMetric.Value != nil {
			compers = *m.Value == *otherMetric.Value
		} else {
			compers = false
		}
	case "counter":
		if m.Delta == nil && otherMetric.Delta == nil {
			compers = true
		} else if m.Delta != nil && otherMetric.Delta != nil {
			compers = *m.Delta == *otherMetric.Delta
		} else {
			compers = false
		}
	default:
		log.Fatalf("need decribe new type=%s", m.MType)
	}

	return m.ID == otherMetric.ID &&
		m.MType == otherMetric.MType &&
		compers
}
