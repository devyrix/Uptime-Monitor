package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckOnceUp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := &Config{History: 5, Targets: []Target{{Name: "test", URL: srv.URL}}}
	s := NewStore(cfg)

	checkOnce(srv.Client(), s, cfg.Targets[0])

	snap := s.Snapshot()
	if len(snap[0].History) != 1 {
		t.Fatalf("harus ada 1 hasil, dapat %d", len(snap[0].History))
	}
	got := snap[0].History[0]
	if !got.Up {
		t.Fatalf("target harusnya UP, dapat %+v", got)
	}
	if got.Status != 200 {
		t.Fatalf("status harusnya 200, dapat %d", got.Status)
	}
	if got.MS < 0 {
		t.Fatalf("latensi tidak boleh negatif, dapat %d", got.MS)
	}
}

func TestCheckOnceServerErrorJadiDown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &Config{History: 5, Targets: []Target{{Name: "test", URL: srv.URL}}}
	s := NewStore(cfg)

	checkOnce(srv.Client(), s, cfg.Targets[0])

	got := s.Snapshot()[0].History[0]
	if got.Up {
		t.Fatal("HTTP 500 harusnya DOWN")
	}
	if got.Status != 500 || got.Error != "" {
		t.Fatalf("status=500 tanpa error transport, dapat %+v", got)
	}
}

func TestCheckOnceKoneksiGagal(t *testing.T) {
	// URL yang tidak mungkin hidup: server yang sudah kita matikan
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	cfg := &Config{History: 5, Targets: []Target{{Name: "mati", URL: srv.URL}}}
	s := NewStore(cfg)

	checkOnce(srv.Client(), s, cfg.Targets[0])

	got := s.Snapshot()[0].History[0]
	if got.Up {
		t.Fatal("target mati harusnya DOWN")
	}
	if got.Status != 0 || got.Error == "" {
		t.Fatalf("gagal total: status harus 0 dan error terisi, dapat %+v", got)
	}
}
