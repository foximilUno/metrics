package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/foximilUno/metrics/internal/config"
	"github.com/foximilUno/metrics/internal/secure"
	"github.com/foximilUno/metrics/internal/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	gauge   = "gauge"
	counter = "counter"

	//available paths
	updatePath      = "/update"
	batchUpdatePath = "/updates/"
	batchSupports   = "/batchSupport"
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
	m       sync.Mutex
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
		sync.Mutex{},
	}
}

func (mc *collector) WithClient(client *http.Client) *collector {
	mc.client = client
	return mc
}

func (mc *collector) addGauge(name string, value uint64) {
	mc.m.Lock()
	defer mc.m.Unlock()

	mc.getEntity(name, gauge).entityValue = value
}

func (mc *collector) increaseCounter(name string) {
	mc.m.Lock()
	defer mc.m.Unlock()

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

func (mc *collector) Report() error {
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
			return fmt.Errorf("unsupported for report type of metric: %s", v.entityType)
		}

		if len(mc.cfg.Key) > 0 {
			encryptVal, err := secure.EncryptMetric(m, mc.cfg.Key)
			if err != nil {
				return fmt.Errorf("cant encrypt metric %v: %w", m, err)
			}
			m.Hash = encryptVal
		}

		metrics = append(metrics, m)
	}

	var err error
	var b []byte
	if mc.isBatchSupports() {
		b, err = json.Marshal(metrics)
		if err != nil {
			return fmt.Errorf("error while marshalling: %w", err)
		}
		err = mc.doRequest(b, mc.baseURL+batchUpdatePath)
		if err != nil {
			return fmt.Errorf("error while send batch: %w", err)
		}
	} else {
		currentURL := mc.baseURL + updatePath
		for _, m := range metrics {
			b, err = json.Marshal(m)
			if err != nil {
				return fmt.Errorf("error while marshalling: %w", err)
			}
			err = mc.doRequest(b, currentURL)
			if err != nil {
				return fmt.Errorf("error while send batch: %w", err)
			}
		}
	}

	log.Println("Reports ended")
	return nil
}

//isBatchUpdateExists checks that path /updates available
func (mc *collector) isBatchSupports() bool {
	rq, err := http.NewRequest(http.MethodGet, mc.baseURL+batchSupports, nil)
	if err != nil {
		return false
	}
	r, err := mc.client.Do(rq)
	if err != nil {
		return false
	}
	defer r.Body.Close()
	return r.StatusCode == http.StatusOK
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
		return fmt.Errorf("error while send request: %e", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server return error for url %s %d", currentURL, resp.StatusCode)
	}
	return nil
}

func (mc *collector) CollectAdditional() error {
	var err error
	stats, err := mem.VirtualMemory()

	if err != nil {
		return err
	}
	cpuStats, err := cpu.Times(true)
	if err != nil {
		return err
	}
	mc.addGauge("TotalMemory", stats.Total)
	mc.addGauge("FreeMemory", stats.Free)
	//хз как посчититаь utilization поэтому берем Idle
	mc.addGauge("CPUutilization1", uint64(cpuStats[0].Idle))
	return err
}
