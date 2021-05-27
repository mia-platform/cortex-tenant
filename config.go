package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

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

	u, err := url.Parse(cfg.Target)
	if err != nil {
		log.Fatal(err)
	}

	AUTH_USER_PASS := os.Getenv("AUTH_USER_PASS")

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
