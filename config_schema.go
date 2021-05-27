package main

import (
	"time"

	fhu "github.com/valyala/fasthttp/fasthttputil"
)

type config struct {
	Listen      string
	ListenPprof string `yaml:"listen_pprof"`

	Target string

	LogLevel        string `yaml:"log_level"`
	Timeout         time.Duration
	TimeoutShutdown time.Duration `yaml:"timeout_shutdown"`

	Tenant struct {
		Label          string `yaml:"label,omitempty"`
		LabelRemove    bool   `yaml:"label_remove,omitempty"`
		NamespaceLabel string `yaml:"namespace_label,omitempty"`
		BatchSize      int    `yaml:"batch_size,omitempty"`
		QueryInterval  int    `yaml:"query_interval,omitempty"`
		Header         string
		Default        string
	}

	pipeIn  *fhu.InmemoryListener
	pipeOut *fhu.InmemoryListener
}
