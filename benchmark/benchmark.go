package benchmark

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// set true if you want to output debug logs
var Debug = false

// signals for trapping while benchmark
var TrapSignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT}

// Runner ... benchmark runner object
type Runner struct {
	Setup       func(*Worker)
	Teardown    func(*Worker)
	Benchmark   func(*Worker) int
	Duration    time.Duration
	Concurrency int
}

// Worker ... benchmark worker object
type Worker struct {
	ID int
	Stash  map[string]interface{}
}

// Result ... benchmark result
type Result struct {
	Score   int
	Elapsed time.Duration
}

func debug (s string, v ...interface{}) {
	if Debug {
		log.Printf(s, v...)
	}
}

// Run benchmark suite
func (r *Runner) Run() *Result {
	c := r.Concurrency
	log.Printf("starting benchmark: concurrency: %d, time: %s, GOMAXPROCS: %d", c, r.Duration, runtime.GOMAXPROCS(0))

	startCh := make(chan bool, c)
	readyCh := make(chan bool, c)
	stopCh := make(chan bool, c)
	scoreCh := make(chan int, c)
	var wg sync.WaitGroup

	// spawn worker goroutines
	for i := 0; i < c; i++ {
		debug("spwan worker[%d]", i);
		go func(n int) {
			wg.Add(1)
			defer wg.Done()
			score := 0
			worker := &Worker{
				ID: n,
				Stash: make(map[string]interface{}),
			}
			if r.Setup != nil {
				debug("worker[%d] Setup()", n)
				r.Setup(worker)
			}
			readyCh <- true // ready of worker:n
			<-startCh       // notified go benchmark from Runner
			debug("worker[%d] starting Benchmark()", n);
		BENCH:
			for {
				select {
				case <-stopCh:
					scoreCh <- score
					break BENCH
				default:
					score += r.Benchmark(worker)
				}
			}
			debug("worker[%d] done Benchmark() score: %d", n, score);
			if r.Teardown != nil {
				debug("worker[%d] Teardown()", n)
				r.Teardown(worker)
			}
			debug("worker[%d] exit", n)
		}(i)
	}

	// wait for ready of workres
	debug("waiting for all workers finish Setup()");
	for i := 0; i < c; i++ {
		<-readyCh
	}

	// notify "start" to workers
	close(startCh)
	start := time.Now()

	// wait for catching signal or timed out
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, TrapSignals...)
	select {
	case s := <-signalCh:
		switch sig := s.(type) {
		case syscall.Signal:
			log.Printf("Got signal: %s(%d)", sig, sig)
		default:
			debug("timed out")
			break
		}
	case <-time.After(r.Duration):
		break
	}

	// notify "stop" to workers
	close(stopCh)

	// collect scores from workers
	totalScore := 0
	for i := 0; i < c; i++ {
		totalScore += <-scoreCh
	}
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("done benchmark: score %d, elapsed %s = %f / sec\n", totalScore, elapsed, float64(totalScore) / float64(elapsed) * float64(time.Second))

	wg.Wait()
	return &Result{ Score: totalScore, Elapsed: elapsed }
}
