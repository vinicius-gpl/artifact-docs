package apps

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type recordingStore struct {
	mutex    sync.Mutex
	targets  []Target
	statuses map[string]string
	icons    map[string]string
}

func (store *recordingStore) Targets(_ context.Context) ([]Target, error) {
	return store.targets, nil
}

func (store *recordingStore) SetStatus(_ context.Context, id, status string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.statuses[id] = status

	return nil
}

func (store *recordingStore) SetIcon(_ context.Context, id, iconURL string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.icons == nil {
		store.icons = map[string]string{}
	}
	store.icons[id] = iconURL

	return nil
}

func (store *recordingStore) iconOf(id string) string {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	return store.icons[id]
}

func (store *recordingStore) statusOf(id string) string {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	return store.statuses[id]
}

func answering(test *testing.T, status int) string {
	test.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(status)
	}))
	test.Cleanup(server.Close)

	return server.URL
}

// feliz
func TestProbeCountsAnyAnswerBelow500AsOnline(test *testing.T) {
	client := &http.Client{Timeout: time.Second}

	for _, status := range []int{http.StatusOK, http.StatusUnauthorized, http.StatusMethodNotAllowed} {
		if got := probe(context.Background(), client, answering(test, status)); got != statusOnline {
			test.Errorf("answer %d: status = %q, want online (the app is up)", status, got)
		}
	}
}

// triste
func TestProbeCountsServerErrorsAndSilenceAsOffline(test *testing.T) {
	client := &http.Client{Timeout: time.Second}
	silent := httptest.NewServer(http.NotFoundHandler())
	silent.Close()

	if got := probe(context.Background(), client, answering(test, http.StatusBadGateway)); got != statusOffline {
		test.Errorf("502: status = %q, want offline", got)
	}
	if got := probe(context.Background(), client, silent.URL); got != statusOffline {
		test.Errorf("nobody listening: status = %q, want offline", got)
	}
	if got := probe(context.Background(), client, "::not a url"); got != statusOffline {
		test.Errorf("broken address: status = %q, want offline", got)
	}
}

// feliz
func TestMonitorChecksEverythingAtStartAndWhatIsAskedSoon(test *testing.T) {
	store := &recordingStore{
		targets:  []Target{{ID: "up", URL: answering(test, http.StatusOK)}},
		statuses: map[string]string{},
	}
	monitor := NewMonitor(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go monitor.Run(ctx)
	monitor.CheckSoon(Target{ID: "down", URL: answering(test, http.StatusServiceUnavailable)})

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && (store.statusOf("up") == "" || store.statusOf("down") == "") {
		time.Sleep(10 * time.Millisecond)
	}
	if store.statusOf("up") != statusOnline || store.statusOf("down") != statusOffline {
		test.Fatalf("statuses = up:%q down:%q", store.statusOf("up"), store.statusOf("down"))
	}
}
