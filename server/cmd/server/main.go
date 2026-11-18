package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"bashcapital/server/internal/api"
	"bashcapital/server/internal/sim"
	"bashcapital/server/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dataDir := flag.String("data", "", "path to data/ JSON")
	tickHours := flag.Float64("tick-hours", 1, "real hours per tick; 1 tick = 1/24 game month (0 = manual only)")
	devFast := flag.Bool("dev", false, "dev mode: tick every 30s")
	flag.Parse()

	dir := *dataDir
	if dir == "" {
		candidates := []string{"../data", "../../data", "data"}
		for _, c := range candidates {
			if st, err := os.Stat(c); err == nil && st.IsDir() {
				dir, _ = filepath.Abs(c)
				break
			}
		}
	}
	if dir == "" {
		log.Fatal("data directory not found; pass -data")
	}

	cat, err := store.LoadCatalog(dir)
	if err != nil {
		log.Fatal(err)
	}
	engine := sim.NewEngine(cat)

	tickEvery := time.Duration(*tickHours * float64(time.Hour))
	manual := false
	if *devFast {
		tickEvery = 30 * time.Second
		manual = false
		log.Println("DEV mode: auto-tick every 30s")
	}
	if *tickHours == 0 {
		tickEvery = 0
		manual = true
		log.Println("Manual tick mode (POST /api/tick)")
	}

	srv := api.New(engine, tickEvery, manual)
	srv.StartTicker()

	fmt.Printf("BashCapital server on %s (data=%s)\n", *addr, dir)
	log.Fatal(http.ListenAndServe(*addr, srv.Handler()))
}
