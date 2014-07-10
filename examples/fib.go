package main

import (
	"log"
	"time"
	"os"
	"strconv"
	"github.com/fujiwara/parallel-benchmark/benchmark"
)

func main() {
	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Panicf("invalid number", os.Args[1])
	}
	runner := &benchmark.Runner{
		Benchmark: func(w *benchmark.Worker) int {
			fib(n)
			return 1
		},
		Duration: time.Duration(3)*time.Second,
		Concurrency: 4,
	}
	runner.Run()
}

func fib(n int) int {
	if n == 0 {
		return 0;
	}
	if n == 1 {
		return 1;
	}
	return (fib(n - 1) + fib(n - 2));
}
