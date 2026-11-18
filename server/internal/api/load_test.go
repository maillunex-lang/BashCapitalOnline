package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"bashcapital/server/internal/model"
	"bashcapital/server/internal/sim"
	"bashcapital/server/internal/store"

	"github.com/gorilla/websocket"
)

type measuredStore struct {
	*store.DB
	saves    atomic.Int64
	errors   atomic.Int64
	saveTime atomic.Int64
}

func (s *measuredStore) SaveWorld(world *model.World) error {
	s.saves.Add(1)
	began := time.Now()
	err := s.DB.SaveWorld(world)
	s.saveTime.Add(int64(time.Since(began)))
	if err != nil {
		s.errors.Add(1)
	}
	return err
}

type playerLoad struct {
	firmID, token string
	socket        *websocket.Conn
}

type loadSample struct {
	firmID        string
	reads, writes []time.Duration
	err           error
}

type tickReceipt struct {
	at  []time.Time
	err error
}

func TestV2LocalLoad(t *testing.T) {
	if os.Getenv("BASHCAPITAL_LOAD") != "1" {
		t.Skip("set BASHCAPITAL_LOAD=1 to run the local workload")
	}
	ticks := 15
	if override := os.Getenv("BASHCAPITAL_LOAD_TICKS"); override != "" {
		var err error
		ticks, err = strconv.Atoi(override)
		if err != nil || ticks <= 0 {
			t.Fatal("BASHCAPITAL_LOAD_TICKS must be a positive integer")
		}
	}
	for _, players := range []int{10, 50, 100} {
		t.Run(fmt.Sprintf("%d_players", players), func(t *testing.T) { runV2Load(t, players, ticks) })
	}
}

func runV2Load(t *testing.T, players, tickCount int) {
	t.Helper()
	const (
		tickPeriod = 100 * time.Millisecond
		warmup     = 10 * time.Millisecond
	)
	workTime := time.Duration(tickCount) * tickPeriod
	catalog, err := store.LoadCatalog("../../../data")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "load.db")
	db, err := store.OpenDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tracker := &measuredStore{DB: db}
	engine := sim.NewEngine(catalog)
	engine.Store = tracker
	server := New(engine, 0, true, "load-test-local", db)
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()
	client := httpServer.Client()
	client.Timeout = 10 * time.Second

	users := make([]playerLoad, players)
	for index := range users {
		username := fmt.Sprintf("load-player-%d", index)
		var auth struct {
			Token string `json:"session_token"`
		}
		if err := loadRequest(client, httpServer.URL, "/api/auth/register", "", map[string]any{"username": username, "password": "load-test-local"}, &auth); err != nil {
			t.Fatal(err)
		}
		var firm model.Firm
		if err := loadRequest(client, httpServer.URL, "/api/register", auth.Token, map[string]any{"name": username, "industry": "food"}, &firm); err != nil {
			t.Fatal(err)
		}
		conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(httpServer.URL, "http")+"/ws", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		users[index] = playerLoad{firm.ID, auth.Token, conn}
	}
	for deadline := time.Now().Add(time.Second); ; {
		server.wsMu.Lock()
		connected := len(server.clients)
		server.wsMu.Unlock()
		if connected == players {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("registered %d/%d websocket clients", connected, players)
		}
		time.Sleep(time.Millisecond)
	}

	samples := make(chan loadSample, players)
	receipts := make(chan tickReceipt, players)
	start := make(chan struct{})
	workDeadline := time.Now().Add(workTime)
	for _, user := range users {
		go func(user playerLoad) {
			sample := loadSample{firmID: user.firmID}
			<-start
			for len(sample.writes) == 0 || time.Now().Before(workDeadline) {
				began := time.Now()
				if sample.err = loadRequest(client, httpServer.URL, "/api/world", user.token, nil, nil); sample.err != nil {
					break
				}
				sample.reads = append(sample.reads, time.Since(began))
				began = time.Now()
				if sample.err = loadRequest(client, httpServer.URL, "/api/action/advertise", user.token, map[string]any{"category": "food", "budget": 1}, nil); sample.err != nil {
					break
				}
				sample.writes = append(sample.writes, time.Since(began))
				time.Sleep(25 * time.Millisecond)
			}
			samples <- sample
		}(user)
		go func(conn *websocket.Conn) {
			receipt := tickReceipt{}
			_ = conn.SetReadDeadline(time.Now().Add(workTime + 15*time.Second))
			for expected := 1; expected <= tickCount; expected++ {
				var message struct {
					Type string `json:"type"`
					Tick int    `json:"tick"`
				}
				if receipt.err = conn.ReadJSON(&message); receipt.err != nil {
					break
				}
				if message.Type != "tick" || message.Tick != expected {
					receipt.err = fmt.Errorf("received %s %d, expected tick %d", message.Type, message.Tick, expected)
					break
				}
				receipt.at = append(receipt.at, time.Now())
			}
			receipts <- receipt
		}(user.socket)
	}

	initialDBStats := db.Stats()
	close(start)
	tickStarts := make([]time.Time, 0, tickCount)
	tickTimes := make([]time.Duration, 0, tickCount)
	schedule := time.Now().Add(warmup)
	for i := 0; i < tickCount; i++ {
		time.Sleep(time.Until(schedule))
		tickStarts = append(tickStarts, time.Now())
		if err := loadRequest(client, httpServer.URL, "/api/tick", "", map[string]any{}, nil); err != nil {
			t.Fatal(err)
		}
		tickTimes = append(tickTimes, time.Since(tickStarts[i]))
		schedule = schedule.Add(tickPeriod)
	}

	var readTimes, writeTimes, deliveryTimes []time.Duration
	writesPerFirm := make(map[string]int, players)
	for range users {
		sample := <-samples
		if sample.err != nil {
			t.Error(sample.err)
		}
		readTimes = append(readTimes, sample.reads...)
		writeTimes = append(writeTimes, sample.writes...)
		writesPerFirm[sample.firmID] = len(sample.writes)
		receipt := <-receipts
		if receipt.err != nil {
			t.Error(receipt.err)
		}
		for index, timestamp := range receipt.at {
			deliveryTimes = append(deliveryTimes, timestamp.Sub(tickStarts[index]))
		}
	}
	if tracker.errors.Load() != 0 || tracker.saves.Load() != int64(len(writeTimes)+tickCount) {
		t.Errorf("SQLite saves=%d errors=%d; expected %d saves", tracker.saves.Load(), tracker.errors.Load(), len(writeTimes)+tickCount)
	}
	stats := db.Stats()
	restored, err := store.OpenDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	saved, err := restored.LoadWorld()
	if err != nil {
		t.Fatal(err)
	}
	if saved.Tick != tickCount {
		t.Errorf("SQLite restored tick %d, expected %d", saved.Tick, tickCount)
	}
	for _, user := range users {
		firm := saved.Firms[user.firmID]
		if firm == nil {
			t.Fatal("SQLite lost a firm")
		}
		// Only advertising changes these asset-free food firms' cash.
		if expected := 115000 - float64(writesPerFirm[user.firmID]); firm.Cash != expected {
			t.Errorf("SQLite restored firm %s with %.0f ₽, expected %.0f ₽", user.firmID, firm.Cash, expected)
		}
	}
	t.Logf("GOMAXPROCS=%d players=%d loads=%d actions=%d SQLite saves=%d failures=%d WS=%d/%d",
		runtime.GOMAXPROCS(0), players, len(readTimes), len(writeTimes), tracker.saves.Load(), tracker.errors.Load(), len(deliveryTimes), players*tickCount)
	t.Logf("SQLite pool waits=%d total_wait=%s mean_save=%s",
		stats.WaitCount-initialDBStats.WaitCount, (stats.WaitDuration - initialDBStats.WaitDuration).Round(time.Millisecond),
		time.Duration(tracker.saveTime.Load()/tracker.saves.Load()).Round(time.Microsecond))
	logLoadLatency(t, "GET world", readTimes)
	logLoadLatency(t, "POST advertise", writeTimes)
	logLoadLatency(t, "POST tick", tickTimes)
	logLoadLatency(t, "tick request to WS receipt", deliveryTimes)
}

func loadRequest(client *http.Client, base, path, token string, body map[string]any, out any) error {
	method := http.MethodGet
	var payload io.Reader
	if body != nil {
		method = http.MethodPost
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, base+path, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Session-Token", token)
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return fmt.Errorf("%s %s: %d %s", method, path, res.StatusCode, body)
	}
	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}
	_, err = io.Copy(io.Discard, res.Body)
	return err
}

func TestBroadcastDropsFullQueueWithoutSkippingOtherReceivers(t *testing.T) {
	s := New(nil, 0, true, "", nil)
	slow, ready := &websocket.Conn{}, &websocket.Conn{}
	slowQueue, readyQueue := make(chan []byte, 1), make(chan []byte, 1)
	s.clients[slow], s.clients[ready] = slowQueue, readyQueue
	slowQueue <- []byte("pending")

	s.broadcast(map[string]any{"type": "tick", "tick": 1})
	if _, exists := s.clients[slow]; exists {
		t.Fatal("a slow receiver with a full queue was retained")
	}
	<-slowQueue
	select {
	case _, open := <-slowQueue:
		if open {
			t.Fatal("slow receiver's queue was not closed")
		}
	default:
		t.Fatal("slow receiver's queue was not closed")
	}
	select {
	case payload := <-readyQueue:
		if string(payload) != `{"tick":1,"type":"tick"}` {
			t.Fatalf("unexpected message %s", payload)
		}
	default:
		t.Fatal("healthy receiver did not get the tick")
	}
}

func logLoadLatency(t *testing.T, group string, values []time.Duration) {
	t.Helper()
	if len(values) == 0 {
		t.Errorf("%s: no latency samples", group)
		return
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	p95 := values[(95*len(values)+99)/100-1]
	t.Logf("%s p95=%s max=%s", group, p95.Round(time.Millisecond), values[len(values)-1].Round(time.Millisecond))
}
