package main

import (
	"context"
	"time"

	"github.com/prometheus/prometheus/prompb"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)


type dispatcher struct {
	clientset *kubernetes.Clientset
	nstenant  map[string]string // namespace_name: tenant_name 
	nstschan map[string]chan *prompb.TimeSeries
	labelName string
	interval int
	proc *processor
}

func newdispatcher(labelName string, interval int, proc *processor) (*dispatcher, error) {
	k := &dispatcher{
		nstenant: make(map[string]string),
		labelName: labelName,
		proc: proc,
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	k.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return k, nil
}
	
func (d *dispatcher) updateMap() (err error) {
	nsList, err := d.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return(err)
	}
	
	for _, ns := range nsList.Items {
		if ns.ObjectMeta.Labels[d.labelName] != "" {
			d.nstenant[ns.ObjectMeta.Name] = ns.ObjectMeta.Labels[d.labelName] 
		} else {
			delete(d.nstenant, ns.ObjectMeta.Name)
		}
	}
	return nil
}

func (d *dispatcher) updateBatchers() {
	for tenant := range d.nstenant {
		_, ok := d.nstschan[tenant]
		if !ok {
			wk := createWorker(tenant, d.proc)
			tschan := make(chan *prompb.TimeSeries)
			go wk.run(tschan)
		}
	}
}
func (d *dispatcher) run() {
	ticker := time.NewTicker(time.Duration(d.interval) * time.Second)
	for ; true; <-ticker.C {
		log.Debug("Call k8s for update ns labels")
		err := d.updateMap()
		if err != nil {
			log.Errorf("Unable to call Api-Server: %s", err)
		}
		d.updateBatchers()
	}
}