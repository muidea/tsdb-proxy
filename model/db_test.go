package model

import (
	"log"
	"testing"
)

// TestConfig test config
func TestConfig(t *testing.T) {
	cfg := ConfigInfo{}

	err := cfg.Load("/Users/rangh/Documents/workspace/dockerComponent/config.json")
	if err != nil {
		t.Errorf("load config failed, err:%s", err.Error())
	}

	log.Print(cfg)
}
