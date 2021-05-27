package main

import (
	"errors"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
	log "github.com/sirupsen/logrus"
	fh "github.com/valyala/fasthttp"
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
	TIMEOUT = 30
)

type timeseries_sender func(proc *processor, tenant string, timeseries []prompb.TimeSeries) (code int, body []byte, err error)

type Worker struct {
	batchsize int
	timeout   int
	tenant    string
	buffer    []prompb.TimeSeries
	sender    timeseries_sender
	proc      *processor
}

func marshal(wr *prompb.WriteRequest) (bufOut []byte, err error) {
	b := make([]byte, wr.Size())
	// Marshal to Protobuf
	if _, err = wr.MarshalTo(b); err != nil {
		return
	}
	// Compress with Snappy
	return snappy.Encode(nil, b), nil
}

func send_timeseries(proc *processor, tenant string, timeseries []prompb.TimeSeries) (int, []byte, error) {
	for c := 0; c < 10; c++ {
		var code int
		var body []byte
		var err error
		req := fh.AcquireRequest()
		resp := fh.AcquireResponse()

		defer func() {
			fh.ReleaseRequest(req)
			fh.ReleaseResponse(resp)
		}()

		wr := prompb.WriteRequest{
			Timeseries: timeseries,
		}

		buf, err := marshal(&wr)
		if err != nil {
			return code, body, err
		}

		req.Header.SetMethod("POST")
		req.Header.Set("Content-Encoding", "snappy")
		req.Header.Set("Content-Type", "application/x-protobuf")
		req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
		req.Header.Set(proc.cfg.Tenant.Header, tenant)
		req.SetRequestURI(proc.cfg.Target)
		req.SetBody(buf)
		if err = proc.cli.DoTimeout(req, resp, proc.cfg.Timeout); err != nil {
			log.Warnf("Error on do timeout: %s", err)
			return code, body, err
		}

		code = resp.Header.StatusCode()
		body = make([]byte, len(resp.Body()))
		copy(body, resp.Body())

		if code == 429 {
			// resp.Header.VisitAll(func (key, value []byte) {
			//     log.Printf("Headers: %v: %v", string(key), string(value))
			// })
			time.Sleep(500)
			continue
		}

		if code != fh.StatusOK {
			log.Errorf("Error on senting writerequest to Cortex: code %d Body %s", code, body)
			return code, body, err
		}
		if code == fh.StatusOK {
			return code, body, err
		}
	}
	return 0, nil, errors.New("finished retry attemps")
}

func createWorker(tenant string, proc *processor) *Worker {
	return &Worker{
		batchsize: proc.cfg.Tenant.BatchSize,
		timeout:   TIMEOUT,
		tenant:    tenant,
		buffer:    make([]prompb.TimeSeries, 0, proc.cfg.Tenant.BatchSize),
		sender:    send_timeseries,
		proc:      proc,
	}
}

func (w *Worker) flush_buffer() {
	cpy := make([]prompb.TimeSeries, len(w.buffer))
	copy(cpy, w.buffer)
	log.Debugf("flushing batcher for tenant: %s", w.tenant)
	go w.sender(w.proc, w.tenant, cpy)
	w.buffer = w.buffer[:0]
}

func (w *Worker) run(tschan <-chan prompb.TimeSeries) {
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
