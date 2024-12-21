package logdb_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/viccon/pulse/clock"
	"github.com/viccon/pulse/logdb"
)

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func copyDir(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dstDir, relPath)

		err = copyFile(path, destPath)
		if err != nil {
			return err
		}

		return nil
	})
}

func TestConcurrentGetSet(t *testing.T) {
	t.Parallel()

	cpus := runtime.NumCPU()
	writeCPUs, readCPUs := cpus/2, cpus/2
	numIterations := 10_000
	db := logdb.NewDB(t.TempDir(), 10, clock.New())

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

func TestUniqueValues(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	err := copyDir("testdata/segments/two", path)
	if err != nil {
		t.Fatal(err)
	}

	mockClock := clock.NewMock(time.Now())
	db := logdb.NewDB(path, 10, mockClock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go db.RunSegmentations(ctx, time.Minute*5)

	values := db.GetAllUnique()
	if len(values) != 11 {
		t.Errorf("expected 11 values, got %d", len(values))
	}
}

func TestAggregation(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	err := copyDir("testdata/segments/two", path)
	if err != nil {
		t.Fatal(err)
	}

	mockClock := clock.NewMock(time.Now())
	db := logdb.NewDB(path, 10, mockClock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go db.RunSegmentations(ctx, time.Minute*5)

	values := db.GetAllUnique()
	if len(values) != 11 {
		t.Errorf("expected 11 values, got %d", len(values))
	}

	aggregatedValues := db.Aggregate()
	if len(aggregatedValues) != 11 {
		t.Errorf("expected 11 values, got %d", len(aggregatedValues))
	}
}

func TestCompaction(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	err := copyDir("testdata/segments/two", path)
	if err != nil {
		t.Fatal(err)
	}

	mockClock := clock.NewMock(time.Now())
	db := logdb.NewDB(path, 10, mockClock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go db.RunSegmentations(ctx, time.Minute*5)

	values := db.GetAllUnique()
	if len(values) != 11 {
		t.Errorf("expected 11 values, got %d", len(values))
	}

	mockClock.Add(time.Minute * 4)
	time.Sleep(time.Millisecond * 250)

	values = db.GetAllUnique()
	if len(values) != 11 {
		t.Errorf("expected 11 values, got %d", len(values))
	}
}

func TestAggregationAfterCompaction(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	err := copyDir("testdata/segments/three", path)
	if err != nil {
		t.Fatal(err)
	}

	mockClock := clock.NewMock(time.Now())
	db := logdb.NewDB(path, 10, mockClock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go db.RunSegmentations(ctx, time.Minute*5)

	values := db.GetAllUnique()
	if len(values) != 11 {
		t.Errorf("expected 11 values, got %d", len(values))
	}

	mockClock.Add(time.Minute * 4)
	time.Sleep(time.Millisecond * 250)

	values = db.GetAllUnique()
	if len(values) != 11 {
		t.Errorf("expected 11 values, got %d", len(values))
	}

	aggregatedValues := db.Aggregate()
	if len(aggregatedValues) != 11 {
		t.Errorf("expected 11 values, got %d", len(aggregatedValues))
	}
}

func TestCompactionWritesAggregation(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	err := copyDir("testdata/segments/two", path)
	if err != nil {
		t.Fatal(err)
	}

	mockClock := clock.NewMock(time.Now())
	db := logdb.NewDB(path, 10, mockClock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go db.RunSegmentations(ctx, time.Minute*5)

	values := db.GetAllUnique()
	if len(values) != 11 {
		t.Errorf("expected 11 values, got %d", len(values))
	}

	mockClock.Add(time.Minute * 4)
	time.Sleep(time.Millisecond * 250)

	values = db.GetAllUnique()
	if len(values) != 11 {
		t.Errorf("expected 11 values, got %d", len(values))
	}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			db.MustSet("key"+strconv.Itoa(i), []byte("value"))
		}
	}

	values = db.GetAllUnique()
	if len(values) != 21 {
		t.Errorf("expected 21 values, got %d", len(values))
	}

	aggregatedValues := db.Aggregate()
	if len(aggregatedValues) != 21 {
		t.Errorf("expected 21 values, got %d", len(aggregatedValues))
	}
	values = db.GetAllUnique()
	if len(values) != 0 {
		t.Errorf("expected 0 values, got %d", len(values))
	}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			db.MustSet("key"+strconv.Itoa(i), []byte("value"))
		}
	}
	values = db.GetAllUnique()
	if len(values) != 10 {
		t.Errorf("expected 10 values, got %d", len(values))
	}
}

func TestAppendingCompactingWritesAggregation(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	err := copyDir("testdata/segments/two", path)
	if err != nil {
		t.Fatal(err)
	}

	mockClock := clock.NewMock(time.Now())
	db := logdb.NewDB(path, 10, mockClock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go db.RunSegmentations(ctx, time.Minute*5)

	values := db.GetAllUnique()
	if len(values) != 11 {
		t.Errorf("expected 11 values, got %d", len(values))
	}

	for i := 0; i < 1000; i++ {
		for j := 0; j < 10; j++ {
			db.MustSet("key"+strconv.Itoa(i), []byte("value"))
		}
	}

	values = db.GetAllUnique()
	if len(values) != 1011 {
		t.Errorf("expected 1011 values, got %d", len(values))
	}

	mockClock.Add(time.Minute * 4)
	time.Sleep(time.Millisecond * 500)

	values = db.GetAllUnique()
	if len(values) != 1011 {
		t.Errorf("expected 1011 values, got %d", len(values))
	}

	aggregatedValues := db.Aggregate()
	if len(aggregatedValues) != 1011 {
		t.Errorf("expected 1011 values, got %d", len(aggregatedValues))
	}
	values = db.GetAllUnique()
	if len(values) != 0 {
		t.Errorf("expected 0 values, got %d", len(values))
	}

	for i := 0; i < 1000; i++ {
		for j := 0; j < 10; j++ {
			db.MustSet("key"+strconv.Itoa(i), []byte("value"))
		}
	}

	values = db.GetAllUnique()
	if len(values) != 1000 {
		t.Errorf("expected 1000 values, got %d", len(values))
	}
}

func TestWritesCompacting(t *testing.T) {
	t.Parallel()

	mockClock := clock.NewMock(time.Now())
	db := logdb.NewDB(t.TempDir(), 10, mockClock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go db.RunSegmentations(ctx, time.Minute*5)

	values := db.GetAllUnique()
	if len(values) != 0 {
		t.Errorf("expected 0 values, got %d", len(values))
	}

	for i := 0; i < 1000; i++ {
		for j := 0; j < 10; j++ {
			db.MustSet("key"+strconv.Itoa(i), []byte("value"))
		}
	}

	values = db.GetAllUnique()
	if len(values) != 1000 {
		t.Errorf("expected 1000 values, got %d", len(values))
	}
}
