package benchmark

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var debug = false

// signals for trapping while benchmark
var TrapSignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT}

// Worker ... benchmark worker interface
type Worker interface {
	Setup()
	Process() int
	Teardown()
}

type funcWorker struct {
	ID            int
	benchmarkFunc func() int
}

func (w *funcWorker) Setup() {
}

func (w *funcWorker) Process() (subscore int) {
	return w.benchmarkFunc()
}

func (w *funcWorker) Teardown() {
}

// Result ... benchmark result
type Result struct {
	Score   int
	Elapsed time.Duration
}

func debugLog(s string, v ...interface{}) {
	if debug {
		log.Printf(s, v...)
	}
}

// RunFunc ... benchmark by function
func RunFunc(benchmarkFunc func() int, duration time.Duration, c int) *Result {
	workers := make([]Worker, c)
	for i := 0; i < c; i++ {
		workers[i] = &funcWorker{ID: i, benchmarkFunc: benchmarkFunc}
	}
	return Run(workers, duration)
}

// Run ... benchmark by workers
func Run(workers []Worker, duration time.Duration) *Result {
	debug = os.Getenv("DEBUG") != ""
	c := len(workers)
	log.Printf("starting benchmark: concurrency: %d, time: %s, GOMAXPROCS: %d", c, duration, runtime.GOMAXPROCS(0))
	startCh := make(chan bool, c)
	readyCh := make(chan bool, c)
	var stopFlag int32
	scoreCh := make(chan int, c)
	var wg sync.WaitGroup

	// spawn worker goroutines
	for i, w := range workers {
		debugLog("spwan worker[%d]", i)
		go func(n int, worker Worker) {
			wg.Add(1)
			defer wg.Done()
			score := 0
			worker.Setup()
			readyCh <- true // ready of worker:n
			<-startCh       // notified go benchmark from Runner
			debugLog("worker[%d] starting Benchmark()", n)
			for atomic.LoadInt32(&stopFlag) == 0 {
				score += worker.Process()
			}
			scoreCh <- score
			debugLog("worker[%d] done Benchmark() score: %d", n, score)
			worker.Teardown()
			debugLog("worker[%d] exit", n)
		}(i, w)
	}

	// wait for ready of workres
	debugLog("waiting for all workers finish Setup()")
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
			log.Printf("interrupted %s", s)
			break
		}
	case <-time.After(duration):
		debugLog("timed out")
		break
	}

	// notify "stop" to workers
	atomic.StoreInt32(&stopFlag, 1)

	// collect scores from workers
	totalScore := 0
	for i := 0; i < c; i++ {
		totalScore += <-scoreCh
	}
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("done benchmark: score %d, elapsed %s = %f / sec\n", totalScore, elapsed, float64(totalScore)/float64(elapsed)*float64(time.Second))

	wg.Wait()
	return &Result{Score: totalScore, Elapsed: elapsed}
}
