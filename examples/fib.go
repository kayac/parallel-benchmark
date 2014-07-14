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
