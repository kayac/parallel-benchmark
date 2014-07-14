package main

import (
	"flag"
	"github.com/kayac/parallel-benchmark/benchmark"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type myWorker struct {
	URL    string
	client *http.Client
}

func (w *myWorker) Setup() {
	w.client = &http.Client{}
}

func (w *myWorker) Teardown() {
}

func (w *myWorker) Process() (subscore int) {
	resp, err := w.client.Get(w.URL)
	if err == nil {
		defer resp.Body.Close()
		_, _ = ioutil.ReadAll(resp.Body)
		if resp.StatusCode == 200 {
			return 1
		}
	} else {
		log.Printf("err: %v, resp: %#v", err, resp)
	}
	return 0
}

func main() {
	var (
		conn     int
		duration int
	)
	flag.IntVar(&conn, "c", 1, "connections to keep open")
	flag.IntVar(&duration, "d", 1, "duration of benchmark")
	flag.Parse()
	url := flag.Args()[0]
	workers := make([]benchmark.Worker, conn)
	for i, _ := range workers {
		workers[i] = &myWorker{URL: url}
	}
	benchmark.Run(workers, time.Duration(duration)*time.Second)
}
