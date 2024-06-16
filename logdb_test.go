package pulse_test

import (
	"runtime"
	"sync"
	"testing"

	"github.com/creativecreature/pulse"
)

func TestConcurrentGetSet(t *testing.T) {
	cpus := runtime.NumCPU()
	writeCPUs, readCPUs := cpus/2, cpus/2
	numIterations := 10_000
	db := pulse.NewDB(t.TempDir())

	wg := sync.WaitGroup{}
	wg.Add(numIterations * (writeCPUs + readCPUs))

	for i := 0; i < writeCPUs; i++ {
		go func() {
			for j := 0; j < numIterations; j++ {
				db.MustSet("key", []byte("value"))
				wg.Done()
			}
		}()
	}

	for i := 0; i < readCPUs; i++ {
		go func() {
			for j := 0; j < numIterations; j++ {
				db.Get("key")
				wg.Done()
			}
		}()
	}

	wg.Wait()
}

// func TestAggregation(t *testing.T) {
// 	db := pulse.NewDB("testdata/segments")
// 	values := db.Aggregate()
// 	if len(values) != 0 {
// 		t.Fatalf("expected 0 values, got %d", len(values))
// 	}
// }
