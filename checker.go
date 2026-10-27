package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// StartCheckers menjalankan satu goroutine pengecekan per target.
// Cek pertama dilakukan langsung, lalu tiap cfg.Interval.
// Nilai kembaliannya adalah fungsi stop() untuk menghentikan semuanya.
func StartCheckers(cfg *Config, store *Store) (stop func()) {
	client := &http.Client{Timeout: cfg.Timeout}
	done := make(chan struct{})
	var wg sync.WaitGroup

	for _, t := range cfg.Targets {
		wg.Add(1)
		go func(t Target) {
			defer wg.Done()
			checkOnce(client, store, t) // cek pertama: langsung, tanpa nunggu interval
			ticker := time.NewTicker(cfg.Interval)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					checkOnce(client, store, t)
				}
			}
		}(t)
	}

	return func() {
		close(done)
		wg.Wait()
	}
}

// checkOnce melakukan satu pengecekan terhadap satu target,
// lalu menyimpan hasilnya ke store dan mencatatnya ke log.
func checkOnce(client *http.Client, store *Store, t Target) {
	start := time.Now()
	res := Result{At: start}

	resp, err := client.Get(t.URL)
	res.MS = time.Since(start).Milliseconds()
	if err != nil {
		res.Error = err.Error()
	} else {
		resp.Body.Close() // wajib: mencegah kebocoran koneksi
		res.Status = resp.StatusCode
		res.Up = resp.StatusCode < 400 // UP = kode HTTP < 400
	}

	store.Record(t.URL, res)
	if res.Up {
		log.Printf("[UP]   %s (%d ms, HTTP %d)", t.Name, res.MS, res.Status)
	} else {
		log.Printf("[DOWN] %s (%s)", t.Name, res.Error)
	}
}
