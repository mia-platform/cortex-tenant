package main

import (
	"fmt"
	"time"

	"github.com/prometheus/prometheus/prompb"
)

// func handle(){
//     arriva la chiamata con dentro delle timeseries
//     divide le timeseries per tenant grazie alla mappa fatta dal poller
//     aggiunge alla coda corrispondente
//     for ts in timeseries:
//         append(ts, queue[tenant])
// }

// func tenantAdmin() {
//     in background lancia un worker per ogni coda delle timeseries ( una coda per ogni tenant )
//     OPZIONALE: se una coda rimane vuota per tot sec cancella il worker
// }

//lint:file-ignore U1000 TDD

const (
    BATCHSIZE = 100
    TIMEOUT = 30
)

type timeseries_sender func(timeseries []*prompb.TimeSeries) error

type Worker struct {
    batchsize int
    timeout int
    tenant string
    buffer []*prompb.TimeSeries
    sender timeseries_sender
    proc *processor
}

func send_timeseries(timeseries []*prompb.TimeSeries) error {
    fmt.Println("mando timeseries veramente")
    return nil
}

func createWorker(tenant string, proc *processor) *Worker{
    return &Worker{
        batchsize: BATCHSIZE,
        timeout: TIMEOUT,
        tenant: tenant,
        buffer: make([]*prompb.TimeSeries, 0, BATCHSIZE),
        sender: send_timeseries,
        proc: proc,
    }
}

func (w *Worker) flush_buffer() {
    cpy := make([]*prompb.TimeSeries, len(w.buffer))
    copy(cpy, w.buffer)
    go w.sender(cpy)
    w.buffer = w.buffer[:0]
}

func (w *Worker) run(tschan <-chan *prompb.TimeSeries){
    // quando le timeseries sono tot o è passato tot tempo le manda
    for {
        select {
            case ts, more := <-tschan:
                if !more {
                    w.flush_buffer()
                    return
                }
                w.buffer = append(w.buffer, ts)
                if len(w.buffer) == w.batchsize {
                    w.flush_buffer()
                }
            case <-time.After(time.Duration(w.timeout) * time.Second):
                if len(w.buffer) != 0 {
                    w.flush_buffer()
                }
        }
    }
}

// func k8spoller(){
//     chiama l'apiserver
//     costruisce una mappa di namespace: tenant ( per ogni label)
// }

// main(){
//     lancia in background il k8spoller
//     lancia in background il tenantAdmin
//     lancia in background server http che lancia handle per ogni chiamata che gli arriva
//     aspetta per sempre
// }