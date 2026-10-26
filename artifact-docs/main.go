// Command artifact-docs sobe o hub: a API em /api e a SPA embutida em todo o resto.
//
// Um binário só, sem Node em produção: a VPS tem 1 GB de RAM, e cada processo a menos é
// memória a mais para o resto. Todo o código Go mora em src-go/; aqui só se liga as peças e
// se cuida da vida do processo (subir, parar com Ctrl+C ou SIGTERM do systemd).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/config"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/server"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/web"
)

// shutdownTimeout é quanto as requisições em andamento têm para terminar quando o processo
// recebe ordem de parar. Passou disso, a conexão é cortada.
const shutdownTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           server.New(logger, web.Dist()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", srv.Addr)
		errs <- srv.ListenAndServe()
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	logger.Info("server stopped")

	return nil
}
