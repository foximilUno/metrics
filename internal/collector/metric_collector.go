package collector

import (
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
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
	host string
	port string
	data map[string]*MetricEntity
}

func NewMetricCollector(host string, port string) *collector {
	return &collector{
		host,
		port,
		make(map[string]*MetricEntity),
	}
}

func (mc *collector) addGauge(name string, value uint64) {
	mc.getEntity(name, gauge).entityValue = value
}

func (mc *collector) increaseCounter(name string) {
	mc.getEntity(name, counter).entityValue++
}

func (mc *collector) getEntity(name string, typeentity string) *MetricEntity {
	if _, ok := mc.data[name]; !ok {
		mc.data[name] = &MetricEntity{typeentity, name, 0}
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
	rand.Seed(time.Now().UnixNano())
	mc.addGauge("RandomValue", rand.Uint64())

	mc.increaseCounter("PollCount")

	log.Printf("Poll %s\r\n", strconv.FormatUint(mc.data["PollCount"].entityValue, 10))
}

func (mc *collector) Report() {
	log.Println("Report to server collect data")
	client := http.Client{}

	commonURL := "http://" + mc.host + ":" + mc.port + "/update"
	for _, v := range mc.data {
		currentURL := commonURL + "/" + v.entityType + "/" + v.entityName + "/" + strconv.FormatUint(v.entityValue, 10)
		req, err := http.NewRequest(http.MethodPost, currentURL, nil)
		if err != nil {
			//TODO what to do)) just logging right now
			log.Println("error while make request", err)
		}
		req.Header.Set("Content-Type", "text/plain")

		resp, err := client.Do(req)

		if err != nil {
			//TODO what to do)) just logging right now
			log.Fatalf("error while send request: %e", err)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Printf("server return error for url %s %d\n", currentURL, resp.StatusCode)
		}
	}
	log.Println("Reports ended")

}
