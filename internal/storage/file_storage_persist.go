package storage

import (
	"encoding/json"
	"fmt"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/types"
	"io"
	"log"
	"os"
)

type filePersist struct {
	filename string
}

func (f *filePersist) Load() (map[string]*types.Metrics, error) {
	metrics := make(map[string]*types.Metrics)
	file, err := os.OpenFile(f.filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	decoder := json.NewDecoder(file)

	var dump *Dump

	if err := decoder.Decode(&dump); err == io.EOF {
		return nil, fmt.Errorf("eof error")
	} else if err != nil {
		log.Fatal("fatal", err)
	}

	log.Printf("loaded %d metrics\n", len(dump.DumpedMetrics))

	for _, v := range dump.DumpedMetrics {
		t := v
		metrics[v.ID] = &t
	}
	return metrics, nil
}

func (f *filePersist) Dump(metrics map[string]*types.Metrics) error {
	log.Println("save to", f.filename)
	file, err := os.OpenFile(f.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)
	if err != nil {
		return err
	}

	metricsArray := &Dump{
		[]types.Metrics{},
	}

	for _, v := range metrics {
		metricsArray.DumpedMetrics = append(metricsArray.DumpedMetrics, *v)
	}

	encoder := json.NewEncoder(file)
	err = encoder.Encode(metricsArray)
	if err != nil {
		return err
	}
	return nil
}

func NewFilePersist(filename string) repositories.Persist {
	return &filePersist{filename: filename}
}
