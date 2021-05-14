package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	fhu "github.com/valyala/fasthttp/fasthttputil"
	"gopkg.in/yaml.v2"
)

type config struct {
	Listen      string
	ListenPprof string `yaml:"listen_pprof"`

	Target string

	LogLevel        string `yaml:"log_level"`
	Timeout         time.Duration
	TimeoutShutdown time.Duration `yaml:"timeout_shutdown"`

	Tenant struct {
		Label       string `yaml:"label,omitempty"`
		LabelRemove bool `yaml:"label_remove,omitempty"`
		NamespaceLabel string `yaml:"namespace_label,omitempty"`
		BatchSize int `yaml:"batch_size,omitempty"`
		QueryInterval int `yaml:"query_interval,omitempty"`
		Header      string
		Default     string
	}

	pipeIn  *fhu.InmemoryListener
	pipeOut *fhu.InmemoryListener
}

func configParse(b []byte) (*config, error) {
	cfg := &config{}
	if err := yaml.UnmarshalStrict(b, cfg); err != nil {
		return nil, errors.Wrap(err, "Unable to parse config")
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	if cfg.Tenant.Header == "" {
		cfg.Tenant.Header = "X-Scope-OrgID"
	}

	if cfg.Tenant.Label == "" {
		cfg.Tenant.Label = "__tenant__"
	}

	AUTH_USER_PASS := os.Getenv("AUTH_USER_PASS")

	u, err := url.Parse(cfg.Target)
	if err != nil {
		log.Fatal(err)
	}

	if AUTH_USER_PASS != "" {
		u.User = url.UserPassword(strings.Split(AUTH_USER_PASS, ":")[0], strings.Split(AUTH_USER_PASS, ":")[1])
	}

	cfg.Target = u.String()

	return cfg, nil
}

func configLoad(file string) (*config, error) {
	y, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read config")
	}

	return configParse(y)
}
