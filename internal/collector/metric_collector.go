package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/foximilUno/metrics/internal/config"
	"github.com/foximilUno/metrics/internal/secure"
	"github.com/foximilUno/metrics/internal/types"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

type MetricEntity struct {
	entityType  string
	entityName  string
	entityValue uint64
}

type collector struct {
	baseURL string
	data    map[string]*MetricEntity
	client  *http.Client
	cfg     *config.Config
}

func NewMetricCollector(cfg *config.Config) *collector {
	var baseURL string
	//contains http/https
	if strings.Contains(cfg.URL, "http") {
		baseURL = cfg.URL
	} else {
		baseURL = "http://" + cfg.URL
	}

	return &collector{
		baseURL,
		make(map[string]*MetricEntity),
		&http.Client{},
		cfg,
	}
}

func (mc *collector) WithClient(client *http.Client) *collector {
	mc.client = client
	return mc
}

func (mc *collector) addGauge(name string, value uint64) {
	mc.getEntity(name, gauge).entityValue = value
}

func (mc *collector) increaseCounter(name string) {
	mc.getEntity(name, counter).entityValue++
}

func (mc *collector) getEntity(name string, typeEntity string) *MetricEntity {
	if _, ok := mc.data[name]; !ok {
		mc.data[name] = &MetricEntity{typeEntity, name, 0}
	}
	return mc.data[name]
}

func (mc *collector) Collect() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	mc.addGauge("Alloc", stats.Alloc)
	mc.addGauge("BuckHashSys", stats.BuckHashSys)
	mc.addGauge("Frees", stats.Frees)
	mc.addGauge("GCCPUFraction", uint64(stats.GCCPUFraction))
	mc.addGauge("GCSys", stats.GCSys)
	mc.addGauge("HeapAlloc", stats.HeapAlloc)
	mc.addGauge("HeapIdle", stats.HeapIdle)
	mc.addGauge("HeapInuse", stats.HeapInuse)
	mc.addGauge("HeapObjects", stats.HeapObjects)
	mc.addGauge("HeapReleased", stats.HeapReleased)
	mc.addGauge("HeapSys", stats.HeapSys)
	mc.addGauge("LastGC", stats.LastGC)
	mc.addGauge("Lookups", stats.Lookups)
	mc.addGauge("MCacheInuse", stats.MCacheInuse)
	mc.addGauge("MCacheSys", stats.MCacheSys)
	mc.addGauge("MSpanInuse", stats.MSpanInuse)
	mc.addGauge("MSpanSys", stats.MSpanSys)
	mc.addGauge("Mallocs", stats.Mallocs)
	mc.addGauge("NextGC", stats.NextGC)
	mc.addGauge("NumForcedGC", uint64(stats.NumForcedGC))
	mc.addGauge("NumGC", uint64(stats.NumGC))
	mc.addGauge("OtherSys", stats.OtherSys)
	mc.addGauge("PauseTotalNs", stats.PauseTotalNs)
	mc.addGauge("StackInuse", stats.StackInuse)
	mc.addGauge("StackSys", stats.StackSys)
	mc.addGauge("Sys", stats.Sys)
	mc.addGauge("RandomValue", rand.Uint64())
	mc.addGauge("TotalAlloc", stats.TotalAlloc)
	mc.increaseCounter("PollCount")

	log.Printf("Poll %s\r\n", strconv.FormatUint(mc.data["PollCount"].entityValue, 10))
}

func (mc *collector) Report() {
	log.Println("Report to server collect data")

	var metrics []*types.Metrics

	for _, v := range mc.data {

		m := &types.Metrics{
			ID:    v.entityName,
			MType: v.entityType,
		}

		switch v.entityType {
		case gauge:
			newVal := float64(v.entityValue)
			m.Value = &newVal
		case counter:
			newVal := int64(v.entityValue)
			m.Delta = &newVal
		default:
			panic(fmt.Sprintf("unsupported for report type of metric: %s", v.entityType))
		}
		if len(mc.cfg.Key) > 0 {
			encryptVal, err := secure.EncryptMetric(m, mc.cfg.Key)
			if err != nil {
				log.Fatalf("cant encrypt metric %v: %e", m, err)
			}
			m.Hash = encryptVal
		}

		metrics = append(metrics, m)
	}

	if mc.isBatchUpdateExists(mc.baseURL + "/updates") {
		b, err := json.Marshal(metrics)
		if err != nil {
			//TODO what to do)) just logging right now
			log.Println("error while marshalling", err)
		}
		if err := mc.doRequest(b, mc.baseURL+"/updates"); err != nil {
			log.Printf("error: %e", err)
			return
		}
	} else {
		currentURL := mc.baseURL + "/update"
		for _, m := range metrics {
			b, err := json.Marshal(m)
			if err != nil {
				//TODO what to do)) just logging right now
				log.Println("error while marshalling", err)
			}
			if err := mc.doRequest(b, currentURL); err != nil {
				log.Printf("error: %e", err)
				return
			}
		}
	}
	log.Println("Reports ended")
}

//isBatchUpdateExists checks that path /updates available
func (mc *collector) isBatchUpdateExists(checkURL string) bool {
	rq, err := http.NewRequest(http.MethodPost, checkURL, nil)
	if err != nil {
		return false
	}
	r, err := mc.client.Do(rq)
	if err != nil {
		return false
	}
	return r.StatusCode != http.StatusNotFound
}

func (mc *collector) doRequest(b []byte, currentURL string) error {

	req, err := http.NewRequest(http.MethodPost, currentURL, bytes.NewBuffer(b))

	if err != nil {
		//TODO what to do)) just logging right now
		return fmt.Errorf("error while make request: %e", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := mc.client.Do(req)

	if err != nil {
		//TODO what to do)) just logging right now
		return fmt.Errorf("error while send request: %e\n", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server return error for url %s %d\n", currentURL, resp.StatusCode)
	}
	return nil
}
