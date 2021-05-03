package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
)

func Test_newWorker(t *testing.T){
	wk := createWorker("testA")
	assert.Equal(t, "testA", wk.tenant)
}

func Test_mockSend(t *testing.T){
	var worked string
	mc := func(timeseries []*prompb.TimeSeries) error {worked="YESSSS";return nil;}
	wk := createWorker("testA")
	wk.sender = mc
	wk.batchsize = 1

	tschan := make(chan *prompb.TimeSeries)
	go wk.run(tschan)
	tschan <- &prompb.TimeSeries{Samples: testTS1.Samples}
	assert.Equal(t, "YESSSS", worked)
}

func Test_flush_buffer(t *testing.T){
	mc := func(timeseries []*prompb.TimeSeries) error {return nil;}
	wk := createWorker("testA")
	wk.sender = mc
	ts := &prompb.TimeSeries{Samples: testTS1.Samples}
	assert.Empty(t, wk.buffer)
	wk.buffer = []*prompb.TimeSeries{ts, ts, ts, ts, ts}
	wk.flush_buffer()
	assert.Empty(t, wk.buffer)
}


func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func Test_flushing_times(t *testing.T){
    var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	mc := func(timeseries []*prompb.TimeSeries) error {waitGroup.Done();return nil;}
	wk := createWorker("testA")
	wk.sender = mc
	wk.batchsize = 5

	tschan := make(chan *prompb.TimeSeries)
	go wk.run(tschan)
	for n := 0; n < 11; n++ {
		tschan <- &prompb.TimeSeries{Samples: testTS1.Samples}
	}
	if waitTimeout(&waitGroup, time.Second) {
		t.Log("wait timeout - expected exactly 2 flush")
		t.FailNow()
	}
}

func Test_flushOnTimeout(t *testing.T){
	var worked string
	mc := func(timeseries []*prompb.TimeSeries) error {worked="YESSSS";return nil;}
	wk := createWorker("testA")
	wk.sender = mc
	wk.batchsize = 2
	wk.timeout = 3

	tschan := make(chan *prompb.TimeSeries)
	go wk.run(tschan)
	tschan <- &prompb.TimeSeries{Samples: testTS1.Samples}
	time.Sleep(4 * time.Second)
	assert.Equal(t, "YESSSS", worked)
}

func run_batchsize(batchsize int, b *testing.B) {
	b.ReportAllocs()
	var ops uint64
	mc := func(timeseries []*prompb.TimeSeries) error {atomic.AddUint64(&ops, uint64(len(timeseries)));return nil;}
	wk := createWorker("testA")
	wk.sender = mc
	wk.batchsize = batchsize
	tschan := make(chan *prompb.TimeSeries)
	go wk.run(tschan)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tschan <- &prompb.TimeSeries{Samples: testTS1.Samples}
	}
	// b.Logf("%s: ops: %d\n", b.Name(), ops)
}

func BenchmarkRunCalculate10(b *testing.B)     { run_batchsize(10, b) }
func BenchmarkRunCalculate100(b *testing.B)     { run_batchsize(100, b) }
func BenchmarkRunCalculate500(b *testing.B)     { run_batchsize(500, b) }
func BenchmarkRunCalculate1000(b *testing.B)     { run_batchsize(1000, b) }
func BenchmarkRunCalculate2000(b *testing.B)     { run_batchsize(2000, b) }