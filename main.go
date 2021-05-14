package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	
	var disp *dispatcher
	proc := newProcessor(*cfg, disp)

	if cfg.Tenant.NamespaceLabel != "" {
		disp, err = newdispatcher(cfg.Tenant.NamespaceLabel, cfg.Tenant.QueryInterval, proc)
		if err != nil {
			log.Fatalf("Unable to create k8s poller: %s", err)
		}
		go disp.run()
	}
	proc.disp = disp // FIXME please

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
