parallel-benchmark
=============

A benchmarking framework to running parallel functions for Go.

Usage
-----

Functional interface.
`RunFunc(func() int, duration time.Duration, concurrency int)`

```go
import (
	"github.com/kayac/parallel-benchmark/benchmark"
)

func main() {
	benchmark.RunFunc(
		func() (subscore int) {
			// Your code for benchmarking
			return 1  // return sub-score at executed once
		},
		time.Duration(3) * time.Second, // duration of benchmarking
		10, // number of parallel goroutines
	)
}
```

Your specified benchmark wokers (running individual goroutines).
`Run(workers []benchmark.Workers, duration time.Duration`

```go
type myWorker struct {
}

func (w *myWorker) Setup() {
	// setup your worker before benchmarking
}

func (w *myWorker) Teardown() {
	// teardown your worker after benchmarking
}

func (w *myWorker) Process() (subscore int) {
	// Your code for benchmarking
	return 1  // return sub-score by executed once
}

func main() {
	workers := make([]benchmark.Worker, 10)
	for i, _ := range workers {
		workers[i] = &myWorker{}
	}
	benchmark.Run(workers, time.Duration(10)*time.Second)
}
```

Example of fibonacci benchmark
------

* A example of calculate `fib(30)` in parallel 10 goroutines.
* `fib(30)` calculated at once then score = 1.

```
$ go run examples/fib.go 30
2014/07/14 12:40:40 starting benchmark: concurrency: 10, time: 3s, GOMAXPROCS: 1
2014/07/14 12:40:43 done benchmark: score 371, elapsed 3.304056579s = 112.286213 / sec
2014/07/14 12:40:43 &benchmark.Result{Score:371, Elapsed:3304056579}

$ GOMAXPROCS=2 go run examples/fib.go 30
2014/07/14 12:40:48 starting benchmark: concurrency: 10, time: 3s, GOMAXPROCS: 2
2014/07/14 12:40:51 done benchmark: score 600, elapsed 3.199295602s = 187.541282 / sec
2014/07/14 12:40:51 &benchmark.Result{Score:600, Elapsed:3199295602}

$ GOMAXPROCS=4 go run examples/fib.go 30
2014/07/14 12:40:55 starting benchmark: concurrency: 10, time: 3s, GOMAXPROCS: 4
2014/07/14 12:40:58 done benchmark: score 1014, elapsed 3.043525503s = 333.166257 / sec
2014/07/14 12:40:58 &benchmark.Result{Score:1014, Elapsed:3043525503}
```

```go
package main

import (
	"github.com/kayac/parallel-benchmark/benchmark"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Panicf("invalid number", os.Args[1])
	}
	result := benchmark.RunFunc(
		func() (subscore int) {
			fib(n)
			return 1
		},
		time.Duration(3) * time.Second,
		10,
	)
	log.Printf("%#v", result)
}

func fib(n int) int {
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	return (fib(n-1) + fib(n-2))
}
```

Example of http GET benchmark (like ApacheBench, wrk, etc...)
------

```
$ GOMAXPROCS=4 go run examples/httpbench.go -c 10 -d 3 http://127.0.0.1:8080/
2014/07/14 12:47:04 starting benchmark: concurrency: 10, time: 3s, GOMAXPROCS: 4
2014/07/14 12:47:07 done benchmark: score 12920, elapsed 3.002256618s = 4303.429601 / sec
```

```go
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
```

Author
------

Fujiwara Shunichiro <kayac.shunichiro@gmail.com>

LICENCE
-------

The MIT License (MIT)


