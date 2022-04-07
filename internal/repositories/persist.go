package repositories

import "github.com/foximilUno/metrics/internal/types"

type Persist interface {
	Load() (map[string]*types.Metrics, error)
	Dump(map[string]*types.Metrics) error
}
