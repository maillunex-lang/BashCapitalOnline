package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type blockedActionWriter struct {
	*httptest.ResponseRecorder
	writing chan struct{}
	release chan struct{}
}

func (w *blockedActionWriter) Write(p []byte) (int, error) {
	close(w.writing)
	<-w.release
	return w.ResponseRecorder.Write(p)
}

func TestActionResponseCanWaitWithoutLockingWorld(t *testing.T) {
	p := newEconomyPlaytest(t, "food")
	w := &blockedActionWriter{httptest.NewRecorder(), make(chan struct{}), make(chan struct{})}
	req := httptest.NewRequest(http.MethodPost, "/api/action/advertise", strings.NewReader(`{"category":"food","budget":5}`))
	req.Header.Set("X-Session-Token", p.token)
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.handler.ServeHTTP(w, req)
	}()
	defer func() {
		close(w.release)
		<-done
	}()

	select {
	case <-w.writing:
	case <-time.After(2 * time.Second):
		t.Fatal("action never reached its response")
	}
	health := make(chan int, 1)
	go func() {
		probe := httptest.NewRecorder()
		p.handler.ServeHTTP(probe, httptest.NewRequest(http.MethodGet, "/health", nil))
		health <- probe.Code
	}()
	select {
	case status := <-health:
		if status != http.StatusOK {
			t.Fatalf("health probe returned %d", status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("waiting for the action response blocked world readers")
	}
}
