package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// path config bisa diberikan sebagai argumen pertama:
	//   go run .            → pakai targets.yaml
	//   go run . lain.yaml  → pakai lain.yaml
	cfgPath := "targets.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	log.Printf("config: %d target, interval %s, timeout %s, history %d",
		len(cfg.Targets), cfg.Interval, cfg.Timeout, cfg.History)

	store := NewStore(cfg)
	stop := StartCheckers(cfg, store)
	defer stop()

	engine := html.New("./templates", ".html")
	app := fiber.New(fiber.Config{Views: engine})

	app.Get("/", func(c *fiber.Ctx) error {
		targets := store.Snapshot()
		up := 0
		for _, t := range targets {
			if t.Up {
				up++
			}
		}
		return c.Render("index", fiber.Map{
			"Targets":   targets,
			"UpCount":   up,
			"DownCount": len(targets) - up,
			"Total":     len(targets),
		})
	})
	app.Get("/api/status", func(c *fiber.Ctx) error {
		return c.JSON(store.Snapshot())
	})
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// server jalan di goroutine sendiri; main menunggu sinyal stop
	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()
	log.Println("dashboard: http://localhost:3000")

	// graceful shutdown: tunggu Ctrl+C (SIGINT) atau docker stop (SIGTERM)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Println("shutting down...")
	_ = app.ShutdownWithTimeout(5 * time.Second)
}
