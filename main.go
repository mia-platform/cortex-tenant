package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	version = "0.0.0"
)

func main() {

	cfgFile := flag.String("config", "", "Path to a config file")
	flag.Parse()

	if *cfgFile == "" {
		log.Fatalf("Config file required")
	}

	cfg, err := configLoad(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.ListenPprof != "" {
		go func() {
			if err := http.ListenAndServe(cfg.ListenPprof, nil); err != nil {
				log.Fatalf("Unable to listen on %s: %s", cfg.ListenPprof, err)
			}
		}()
	}

	if cfg.LogLevel != "" {
		lvl, err := log.ParseLevel(cfg.LogLevel)
		if err != nil {
			log.Fatalf("Unable to parse log level: %s", err)
		}

		log.SetLevel(lvl)
	}


	var k8s *k8snspoller
	if cfg.Tenant.NamespaceLabel != "" {
		k8s, err = newK8snspoller(cfg.Tenant.NamespaceLabel)
		if err != nil {
			log.Fatalf("Unable to create k8s Ns Poller: %s", err)
		}
		
		go func() {
			for range time.Tick(time.Duration(cfg.Tenant.QueryInterval) * time.Second ) {
				log.Debug("Call k8s for update ns labels")
				k8s.updateMap()
			}
			}()
	}

	proc := newProcessor(*cfg, *k8s)


	if err = proc.run(); err != nil {
		log.Fatalf("Unable to start: %s", err)
	}

	log.Warnf("Listening on %s", cfg.Listen)
	log.Warnf("Started v%s", version)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, os.Interrupt)
	<-ch

	log.Warn("Shutting down, draining requests")
	if err = proc.close(); err != nil {
		log.Errorf("Error during shutdown: %s", err)
	}

	log.Warnf("Finished")
}
