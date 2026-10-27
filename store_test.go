package main

import (
	"testing"
	"time"
)

func testStore() *Store {
	return NewStore(&Config{
		History: 2,
		Targets: []Target{{Name: "x", URL: "http://x"}},
	})
}

func TestRecordDanSnapshot(t *testing.T) {
	s := testStore()
	s.Record("http://x", Result{At: time.Now(), Up: true, MS: 12, Status: 200})
	s.Record("http://x", Result{At: time.Now(), Up: false, MS: 99, Status: 500})

	snap := s.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("snapshot harus 1 target, dapat %d", len(snap))
	}
	if len(snap[0].History) != 2 {
		t.Fatalf("history harus 2, dapat %d", len(snap[0].History))
	}
	if snap[0].Up {
		t.Fatal("status terakhir harusnya DOWN (hasil kedua)")
	}
	if snap[0].LastMS != 99 {
		t.Fatalf("last_ms harusnya 99, dapat %d", snap[0].LastMS)
	}
}

func TestHistoryTerpotong(t *testing.T) {
	s := testStore() // maxHist = 2
	s.Record("http://x", Result{At: time.Now(), Up: true, MS: 1, Status: 200})
	s.Record("http://x", Result{At: time.Now(), Up: true, MS: 2, Status: 200})
	s.Record("http://x", Result{At: time.Now(), Up: false, MS: 3, Status: 500})

	hist := s.Snapshot()[0].History
	if len(hist) != 2 {
		t.Fatalf("history harus terpotong jadi 2, dapat %d", len(hist))
	}
	if hist[0].MS != 2 {
		t.Fatalf("hasil pertama harusnya yang kedua (ms=2), dapat %d", hist[0].MS)
	}
}

func TestSnapshotSalinan(t *testing.T) {
	s := testStore()
	s.Record("http://x", Result{At: time.Now(), Up: true, MS: 1, Status: 200})

	snap := s.Snapshot()
	snap[0].History[0].MS = 9999 // ubah salinan

	if s.Snapshot()[0].History[0].MS == 9999 {
		t.Fatal("Snapshot harus mengembalikan salinan, bukan referensi internal")
	}
}

func TestRecordURLAsing(t *testing.T) {
	s := testStore()
	s.Record("http://bukan-target", Result{At: time.Now(), Up: true}) // tidak boleh panic
}
