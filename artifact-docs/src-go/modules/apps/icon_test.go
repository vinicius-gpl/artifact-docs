package apps

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// site sobe um servidor com a página e o /favicon.ico que o teste pedir.
func site(test *testing.T, page string, favicon bool) *httptest.Server {
	test.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(writer http.ResponseWriter, _ *http.Request) {
		if !favicon {
			http.NotFound(writer, nil)
			return
		}
		writer.Header().Set("Content-Type", "image/x-icon")
		_, _ = writer.Write([]byte{0, 0, 1, 0})
	})
	mux.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(writer, page)
	})
	server := httptest.NewServer(mux)
	test.Cleanup(server.Close)

	return server
}

func quickClient() *http.Client {
	return &http.Client{Timeout: time.Second}
}

// feliz
func TestDiscoverIconPrefersTheSVGAndResolvesItAgainstThePage(test *testing.T) {
	server := site(test, `<html><head>
		<link rel="stylesheet" href="/app.css">
		<link rel="icon" sizes="32x32" href="/favicon-32.png">
		<LINK REL='icon' type='image/svg+xml' href='img/logo.svg?v=1&amp;x=2'>
		<link rel="apple-touch-icon" href="/apple.png">
	</head></html>`, true)

	got := discoverIcon(context.Background(), quickClient(), server.URL+"/app/")

	if want := server.URL + "/app/img/logo.svg?v=1&x=2"; got != want {
		test.Fatalf("icon = %q, want %q", got, want)
	}
}

func TestBestIconPicksTheLargestPNGWhenThereIsNoSVG(test *testing.T) {
	base, _ := url.Parse("https://exemplo.com/")
	document := `<link rel="shortcut icon" href="/16.png" sizes="16x16">
		<link rel="icon" href="https://cdn.exemplo.com/192.png" sizes="192x192">`

	if got := bestIcon(document, base); got != "https://cdn.exemplo.com/192.png" {
		test.Fatalf("icon = %q, want the 192 px one", got)
	}
}

func TestDiscoverIconFallsBackToTheFaviconOfTheOrigin(test *testing.T) {
	server := site(test, `<html><head><title>sem ícone</title></head></html>`, true)

	if got := discoverIcon(context.Background(), quickClient(), server.URL+"/login"); got != server.URL+"/favicon.ico" {
		test.Fatalf("icon = %q, want the /favicon.ico", got)
	}
}

func TestMonitorSavesTheIconOnlyForWhoDoesNotHaveOne(test *testing.T) {
	server := site(test, `<link rel="icon" href="/logo.svg">`, false)
	store := &recordingStore{
		targets: []Target{
			{ID: "new", URL: server.URL},
			{ID: "known", URL: server.URL, IconURL: "https://exemplo.com/antigo.png"},
		},
		statuses: map[string]string{},
	}
	monitor := NewMonitor(store, slog.New(slog.NewTextHandler(io.Discard, nil)))

	monitor.checkAll(context.Background())

	if got := store.iconOf("new"); got != server.URL+"/logo.svg" {
		test.Errorf("new app icon = %q, want the page's svg", got)
	}
	if got := store.iconOf("known"); got != "" {
		test.Errorf("app that already had an icon was looked up again: %q", got)
	}
}

// triste
func TestDiscoverIconGivesUpWhenNothingIsAnImage(test *testing.T) {
	server := site(test, `<link rel="icon" href="javascript:alert(1)"><link rel="icon" href="data:image/png;base64,AAAA">`, false)

	if got := discoverIcon(context.Background(), quickClient(), server.URL); got != "" {
		test.Fatalf("icon = %q, want none (only unsafe hrefs and no favicon)", got)
	}
}

func TestFallbackFaviconRefusesAPageAnsweredAsFavicon(test *testing.T) {
	// Muito site responde 200 com HTML para qualquer caminho: isso não é ícone.
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(writer, "<html></html>")
	}))
	test.Cleanup(server.Close)

	if got := discoverIcon(context.Background(), quickClient(), server.URL); got != "" {
		test.Fatalf("icon = %q, want none", got)
	}
}

func TestMonitorDoesNotLookForTheIconOfAnAppThatIsDown(test *testing.T) {
	store := &recordingStore{
		targets:  []Target{{ID: "down", URL: answering(test, http.StatusBadGateway)}},
		statuses: map[string]string{},
	}
	monitor := NewMonitor(store, slog.New(slog.NewTextHandler(io.Discard, nil)))

	monitor.checkAll(context.Background())

	if got := store.iconOf("down"); got != "" {
		test.Fatalf("icon of an app that is down = %q, want none", got)
	}
}
