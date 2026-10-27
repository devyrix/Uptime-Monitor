package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTmp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "targets.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfigValid(t *testing.T) {
	path := writeTmp(t, `
interval: 10s
targets:
  - name: Contoh
    url: https://contoh.com
`)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Interval != 10*time.Second {
		t.Fatalf("interval = %v, harusnya 10s", cfg.Interval)
	}
	if cfg.Timeout != 5*time.Second { // nilai bawaan
		t.Fatalf("timeout = %v, harusnya bawaan 5s", cfg.Timeout)
	}
	if cfg.History != 20 { // nilai bawaan
		t.Fatalf("history = %d, harusnya bawaan 20", cfg.History)
	}
	if len(cfg.Targets) != 1 || cfg.Targets[0].Name != "Contoh" {
		t.Fatalf("targets = %+v", cfg.Targets)
	}
}

func TestLoadConfigTanpaTarget(t *testing.T) {
	path := writeTmp(t, `targets: []`)
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("harusnya error kalau tidak ada target")
	}
}

func TestLoadConfigTargetTanpaURL(t *testing.T) {
	path := writeTmp(t, `
targets:
  - name: Rusak
`)
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("harusnya error kalau target tanpa url")
	}
}
