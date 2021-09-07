package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DuplicateToDefault(t *testing.T) {
	cfg, err := configLoad("config.yml")
	if err != nil {
		t.FailNow()
	}
	assert.Equal(t, cfg.Tenant.DuplicateToDefault, false)
}
