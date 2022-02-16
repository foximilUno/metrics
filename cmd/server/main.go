package main

import "github.com/foximilUno/metrics/internal/server"

func main() {

	server.NewMetricServer().ListenAndServe()

}
