package main

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)


type k8snspoller struct {
	clientset *kubernetes.Clientset
	nstenant  map[string]string
	labelName string
}

func newK8snspoller(labelName string) (*k8snspoller, error) {
	k := &k8snspoller{
		nstenant: make(map[string]string),
		labelName: labelName,
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

	k.updateMap()
	return k, nil
}
	
func (p *k8snspoller) updateMap() (err error) {
	nsList, err := p.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return(err)
	}
	
	for _, ns := range nsList.Items {
		if ns.ObjectMeta.Labels[p.labelName] != "" {
			p.nstenant[ns.ObjectMeta.Name] = ns.ObjectMeta.Labels[p.labelName] 
		} else {
			delete(p.nstenant, ns.ObjectMeta.Name)
		}
	}
	return nil
}