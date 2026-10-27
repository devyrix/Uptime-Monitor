package main

import (
	"sync"
	"time"
)

// Result adalah hasil satu kali pengecekan sebuah target.
type Result struct {
	At     time.Time `json:"at"`              // waktu cek
	Up     bool      `json:"up"`              // true = UP
	MS     int64     `json:"ms"`              // latensi milidetik
	Status int       `json:"status"`          // kode HTTP; 0 jika gagal total
	Error  string    `json:"error,omitempty"` // pesan error jika gagal
}

// TargetStatus adalah status kumulatif satu target + riwayatnya.
type TargetStatus struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Up        bool      `json:"up"`         // dari hasil terakhir
	LastMS    int64     `json:"last_ms"`    // latensi terakhir
	CheckedAt time.Time `json:"checked_at"` // waktu cek terakhir
	History   []Result  `json:"history"`
}

// Store menyimpan status semua target di memori.
// Aman dipakai dari beberapa goroutine (checker menulis, handler membaca).
type Store struct {
	mu      sync.RWMutex
	targets map[string]*TargetStatus // kunci: URL target
	order   []string                 // urutan URL sesuai config
	maxHist int
}

func NewStore(cfg *Config) *Store {
	s := &Store{
		targets: make(map[string]*TargetStatus, len(cfg.Targets)),
		order:   make([]string, 0, len(cfg.Targets)),
		maxHist: cfg.History,
	}
	for _, t := range cfg.Targets {
		s.order = append(s.order, t.URL)
		s.targets[t.URL] = &TargetStatus{Name: t.Name, URL: t.URL}
	}
	return s
}

// Record menyimpan satu hasil pengecekan untuk target dengan URL tsb.
func (s *Store) Record(url string, r Result) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ts, ok := s.targets[url]
	if !ok {
		return
	}
	ts.History = append(ts.History, r)
	if len(ts.History) > s.maxHist {
		ts.History = ts.History[len(ts.History)-s.maxHist:]
	}
	ts.Up = r.Up
	ts.LastMS = r.MS
	ts.CheckedAt = r.At
}

// Snapshot mengembalikan salinan status semua target,
// terurut sesuai urutan di config. Salinan = aman dibaca tanpa lock.
func (s *Store) Snapshot() []TargetStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]TargetStatus, 0, len(s.order))
	for _, url := range s.order {
		c := *s.targets[url] // salin struct-nya dulu
		c.History = make([]Result, len(c.History))
		copy(c.History, s.targets[url].History)
		out = append(out, c)
	}
	return out
}
