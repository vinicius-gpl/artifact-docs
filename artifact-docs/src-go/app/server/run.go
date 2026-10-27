package server

import (
	"context"
	"crypto/rand"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/config"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/neonauth"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/storage"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/apps"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/artifacts"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/auth"
)

// shutdownTimeout é quanto as requisições em andamento têm para terminar quando o processo
// recebe ordem de parar. Passou disso, a conexão é cortada.
const shutdownTimeout = 10 * time.Second

// Run sobe o servidor e só volta quando ele para (Ctrl+C ou SIGTERM do systemd).
func Run(logger *slog.Logger, spa fs.FS) error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	deps := Dependencies{Auth: authSettings(cfg), LinkSecret: linkSecret(logger, cfg.LinkSecret)}
	if cfg.IsTest() {
		logger.Warn("APP_ENV=test: every request logs in as the test user, without Neon Auth")
	}
	if cfg.DatabaseURL != "" {
		pool, err := openDatabase(ctx, logger, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer pool.Close()

		deps.Owners = auth.NewOwnerRepository(pool)
		deps.Artifacts = artifacts.NewRepository(pool)
		appRepository := apps.NewRepository(pool)
		deps.Apps = appRepository

		monitor := apps.NewMonitor(appRepository, logger)
		deps.AppChecker = monitor
		go monitor.Run(ctx)
	}
	if cfg.Storage.Configured() {
		bucket, err := storage.Open(storage.Settings(cfg.Storage))
		if err != nil {
			return err
		}
		deps.Files = bucket
	}

	return serve(ctx, logger, &http.Server{
		Addr:              cfg.Addr(),
		Handler:           New(logger, spa, deps),
		ReadHeaderTimeout: 5 * time.Second,
	})
}

// linkSecret é o ARTIFACT_LINK_SECRET, ou um sorteado agora quando ele não foi configurado
// (desenvolvimento): os links privados funcionam, mas param de valer quando o processo
// reinicia.
func linkSecret(logger *slog.Logger, configured string) []byte {
	if configured != "" {
		return []byte(configured)
	}

	logger.Warn("ARTIFACT_LINK_SECRET not set: private artifact links stop working on restart")
	secret := make([]byte, 32)
	_, _ = rand.Read(secret)

	return secret
}

// authSettings liga o porteiro às chaves da Neon. Sem NEON_AUTH_JWKS_URL o login fica
// desligado (Tokens nil) e o servidor sobe assim mesmo: em desenvolvimento a tela usa o mock.
// Com APP_ENV=test, o porteiro deixa entrar o usuário de teste (o e2e), sem Neon Auth.
func authSettings(cfg config.Config) auth.Settings {
	settings := auth.Settings{AllowedEmails: cfg.AllowedEmails, NeonAuthURL: cfg.NeonAuthURL, TestMode: cfg.IsTest()}
	if cfg.NeonAuthJWKSURL == "" {
		return settings
	}

	client := &http.Client{Timeout: 5 * time.Second}
	settings.Tokens = neonauth.NewVerifier(neonauth.NewKeySet(cfg.NeonAuthJWKSURL, client))

	return settings
}

// openDatabase conecta ao Neon e aplica as migrations que faltam. Migration que falha
// derruba a partida: subir com o banco pela metade espalharia o erro por todas as telas.
func openDatabase(ctx context.Context, logger *slog.Logger, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	if err := migrate(ctx, logger, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func migrate(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) error {
	migrations, err := database.Embedded()
	if err != nil {
		return err
	}

	applied, err := database.Migrate(ctx, pool, migrations)
	if err != nil {
		return err
	}
	logger.Info("migrations applied", "names", applied)

	return nil
}

func serve(ctx context.Context, logger *slog.Logger, httpServer *http.Server) error {
	failures := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", httpServer.Addr)
		failures <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-failures:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("server stopped")

	return nil
}
