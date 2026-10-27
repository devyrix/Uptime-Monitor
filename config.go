package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Target adalah satu URL yang dipantau.
type Target struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Config adalah isi penuh targets.yml.
type Config struct {
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	History  int           `yaml:"history"`
	Targets  []Target      `yaml:"targets"`
}

// LoadConfig membaca dan memvalidasi file config.
// Nilai yang tidak diisi diisi bawaan: interval 30s, timeout 5s, history 20.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("baca %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	if cfg.Interval == 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.History == 0 {
		cfg.History = 20
	}
	if len(cfg.Targets) == 0 {
		return nil, fmt.Errorf("%s tidak punya target", path)
	}
	for i, t := range cfg.Targets {
		if t.URL == "" {
			return nil, fmt.Errorf("target #%d (%s) tidak punya url", i+1, t.Name)
		}
	}
	return cfg, nil
}
