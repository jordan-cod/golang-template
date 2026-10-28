package server

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/jordan-cod/golang-template/internal/config"
)

func testConfig(t *testing.T) config.Config {
	t.Helper()
	cfg, err := config.Load(func(string) string { return "" })
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	cfg.ShutdownTimeout = 2 * time.Second
	return cfg
}

func TestServeAndGracefulShutdown(t *testing.T) {
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := New(testConfig(t), discardLogger())
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ctx, ln) }()

	url := "http://" + ln.Addr().String() + "/readyz"
	deadline := time.Now().Add(2 * time.Second)
	for {
		req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("server never became ready: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not return after cancel")
	}

	if srv.health.Ready() {
		t.Error("server still reports ready after shutdown")
	}
}

func TestRunFailsOnBusyPort(t *testing.T) {
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	cfg := testConfig(t)
	cfg.Port = ln.Addr().(*net.TCPAddr).Port

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	if err := New(cfg, discardLogger()).Run(ctx); err == nil {
		t.Fatal("expected error when port is in use")
	}
}

func TestShutdownDrainsInFlightRequests(t *testing.T) {
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	started := make(chan struct{})
	srv := New(testConfig(t), discardLogger())
	srv.http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	})

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ctx, ln) }()

	status := make(chan int, 1)
	go func() {
		req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+ln.Addr().String()+"/slow", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			status <- 0
			return
		}
		_ = resp.Body.Close()
		status <- resp.StatusCode
	}()

	<-started
	cancel()

	if got := <-status; got != http.StatusNoContent {
		t.Errorf("in-flight request status = %d, want 204", got)
	}
	if err := <-done; err != nil {
		t.Errorf("Serve returned error: %v", err)
	}
}
