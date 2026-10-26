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

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/auth"
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

// authFrom liga o porteiro às chaves da Neon. Sem NEON_AUTH_JWKS_URL o login fica desligado
// (Tokens nil) e o servidor sobe assim mesmo: em desenvolvimento a tela usa o modo mock.
func authFrom(cfg config.Config) server.Auth {
	settings := server.Auth{AllowedEmails: cfg.AllowedEmails, NeonAuthURL: cfg.NeonAuthURL}
	if cfg.NeonAuthJWKSURL == "" {
		return settings
	}

	client := &http.Client{Timeout: 5 * time.Second}
	settings.Tokens = auth.NewVerifier(auth.NewKeySet(cfg.NeonAuthJWKSURL, client))

	return settings
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           server.New(logger, web.Dist(), authFrom(cfg)),
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
