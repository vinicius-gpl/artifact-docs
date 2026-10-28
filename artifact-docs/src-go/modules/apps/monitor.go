package apps

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// O monitor do hub (briefing, módulo 4): uma goroutine que confere cada app de tempos em
// tempos e grava "No ar" ou "Fora do ar". Roda dentro do mesmo processo — é um dos motivos de
// o hub precisar de um processo sempre ligado, e não de um serviço que dorme quando ocioso.

const (
	// monitorInterval é de quanto em quanto tempo todos os apps são conferidos.
	monitorInterval = 5 * time.Minute
	// probeTimeout é quanto um app tem para responder antes de contar como fora do ar.
	probeTimeout = 10 * time.Second
	// pendingChecks é quantos "confira logo" cabem na fila. Encheu, o pedido é descartado: a
	// rodada seguinte confere aquele app de qualquer jeito.
	pendingChecks = 16
)

const (
	statusOnline  = "online"
	statusOffline = "offline"
)

// MonitorStore é o que o monitor usa do banco — o Repository em produção.
type MonitorStore interface {
	Targets(ctx context.Context) ([]Target, error)
	SetStatus(ctx context.Context, id, status string) error
	SetIcon(ctx context.Context, id, iconURL string) error
}

// Monitor confere os apps.
type Monitor struct {
	store    MonitorStore
	client   *http.Client
	logger   *slog.Logger
	interval time.Duration
	soon     chan Target
}

// NewMonitor confere os apps de store.
func NewMonitor(store MonitorStore, logger *slog.Logger) *Monitor {
	return &Monitor{
		store:    store,
		client:   &http.Client{Timeout: probeTimeout},
		logger:   logger,
		interval: monitorInterval,
		soon:     make(chan Target, pendingChecks),
	}
}

// Run confere todos agora e depois a cada intervalo, até ctx acabar. Entre as rodadas,
// atende os "confira logo" de quem acabou de cadastrar ou editar um app.
func (monitor *Monitor) Run(ctx context.Context) {
	ticker := time.NewTicker(monitor.interval)
	defer ticker.Stop()

	monitor.checkAll(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			monitor.checkAll(ctx)
		case target := <-monitor.soon:
			monitor.check(ctx, target)
		}
	}
}

// CheckSoon pede a conferência de um app sem esperar por ela: quem cadastrou vê "No ar" em
// segundos, e não na próxima rodada.
func (monitor *Monitor) CheckSoon(target Target) {
	select {
	case monitor.soon <- target:
	default:
	}
}

func (monitor *Monitor) checkAll(ctx context.Context) {
	targets, err := monitor.store.Targets(ctx)
	if err != nil {
		monitor.logger.Error("app monitor: list apps", "error", err)
		return
	}

	for _, target := range targets {
		monitor.check(ctx, target)
	}
}

func (monitor *Monitor) check(ctx context.Context, target Target) {
	status := probe(ctx, monitor.client, target.URL)
	if err := monitor.store.SetStatus(ctx, target.ID, status); err != nil {
		monitor.logger.Error("app monitor: save status", "app", target.ID, "error", err)
	}

	monitor.findIcon(ctx, target, status)
}

// findIcon procura o ícone só de quem ainda não tem e está no ar: baixar o HTML de todo app a
// cada cinco minutos seria tráfego e memória à toa numa VPS de 1 GB, e app fora do ar não
// teria página para ler.
func (monitor *Monitor) findIcon(ctx context.Context, target Target, status string) {
	if target.IconURL != "" || status != statusOnline {
		return
	}

	iconURL := discoverIcon(ctx, monitor.client, target.URL)
	if iconURL == "" {
		return
	}
	if err := monitor.store.SetIcon(ctx, target.ID, iconURL); err != nil {
		monitor.logger.Error("app monitor: save icon", "app", target.ID, "error", err)
	}
}

// probe diz se o app respondeu. Qualquer resposta abaixo de 500 conta como no ar: um 401 ou
// 403 é o app de pé pedindo login, e um 405 é o app de pé recusando o HEAD.
func probe(ctx context.Context, client *http.Client, address string) string {
	request, err := http.NewRequestWithContext(ctx, http.MethodHead, address, nil)
	if err != nil {
		return statusOffline
	}
	request.Header.Set("User-Agent", "artifact-docs-monitor")

	response, err := client.Do(request)
	if err != nil {
		return statusOffline
	}
	_ = response.Body.Close()

	if response.StatusCode >= http.StatusInternalServerError {
		return statusOffline
	}

	return statusOnline
}
