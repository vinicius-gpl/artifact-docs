// Package config lê a configuração do processo a partir do ambiente.
//
// Só variável de ambiente, nada de arquivo: em desenvolvimento quem as carrega é o mise (a
// partir de src-go/.env); em produção, o systemd. O código não sabe de onde elas vieram.
package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// DefaultPort é a porta da API em desenvolvimento — a mesma para a qual o Vite encaminha /api
// (VITE_API_PORT em svelte/.env.example).
const DefaultPort = 3010

// Os ambientes aceitos em APP_ENV.
const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"
)

// Config é tudo o que o processo precisa saber para subir.
type Config struct {
	Port int
	// Env é o APP_ENV. Em "test" (o e2e), o login não passa pelo Neon Auth: toda requisição
	// entra como o usuário de teste, o mesmo do modo mock da tela.
	Env string
	// DatabaseURL é a conexão com o Neon (endpoint pooled). Vazia, o servidor sobe sem
	// banco e as rotas que precisam dele respondem 503 — em desenvolvimento a tela usa o mock.
	DatabaseURL string
	// NeonAuthURL é o endereço do login (Auth → URL do endpoint, no painel da Neon).
	NeonAuthURL string
	// NeonAuthJWKSURL é de onde vêm as chaves que conferem o token. Vazia, o login fica
	// desligado e as rotas protegidas respondem 503 — em desenvolvimento a tela usa o mock.
	NeonAuthJWKSURL string
	// AllowedEmails é quem pode usar o hub, em minúsculas (AUTH_ALLOWED_EMAILS, separados
	// por vírgula).
	AllowedEmails []string
	// Storage é o R2, onde moram os arquivos dos artefatos. Vazio, o servidor sobe sem
	// arquivos: publicar e abrir artefato respondem 503.
	Storage Storage
	// LinkSecret assina os links dos artefatos privados (ARTIFACT_LINK_SECRET). Vazio, o
	// processo sorteia um ao subir — e os links privados abertos param de valer a cada deploy.
	LinkSecret string
}

// Storage são as credenciais do R2 (API S3), do painel da Cloudflare: R2 → Manage API tokens.
type Storage struct {
	// Endpoint é https://<account-id>.r2.cloudflarestorage.com.
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

// Configured diz se o R2 foi configurado.
func (storage Storage) Configured() bool {
	return storage.Endpoint != ""
}

// minLinkSecret é o menor segredo aceito: 32 caracteres aleatórios são 190+ bits.
const minLinkSecret = 32

// Addr é o endereço em que o servidor escuta.
func (cfg Config) Addr() string {
	return fmt.Sprintf(":%d", cfg.Port)
}

// Load monta a Config a partir de getenv (os.Getenv em produção; um mapa nos testes).
//
// Valor inválido é erro, e não volta ao padrão em silêncio: uma porta digitada errada no
// systemd precisa derrubar a partida, e não fazer o servidor subir numa porta que ninguém
// esperava.
func Load(getenv func(string) string) (Config, error) {
	port, err := loadPort(getenv("API_PORT"))
	if err != nil {
		return Config{}, err
	}

	env, err := loadEnv(getenv)
	if err != nil {
		return Config{}, err
	}

	storage, err := loadStorage(getenv)
	if err != nil {
		return Config{}, err
	}

	linkSecret := strings.TrimSpace(getenv("ARTIFACT_LINK_SECRET"))
	if linkSecret != "" && len(linkSecret) < minLinkSecret {
		return Config{}, fmt.Errorf("ARTIFACT_LINK_SECRET must have at least %d characters", minLinkSecret)
	}

	return Config{
		Port:            port,
		Env:             env,
		DatabaseURL:     strings.TrimSpace(getenv("DATABASE_URL")),
		NeonAuthURL:     strings.TrimSpace(getenv("NEON_AUTH_URL")),
		NeonAuthJWKSURL: strings.TrimSpace(getenv("NEON_AUTH_JWKS_URL")),
		AllowedEmails:   splitEmails(getenv("AUTH_ALLOWED_EMAILS")),
		Storage:         storage,
		LinkSecret:      linkSecret,
	}, nil
}

func loadPort(raw string) (int, error) {
	if raw == "" {
		return DefaultPort, nil
	}

	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("API_PORT must be a port between 1 and 65535, got %q", raw)
	}

	return port, nil
}

// splitEmails normaliza a lista: e-mail é comparado sem diferença de maiúscula, e um espaço
// depois da vírgula no systemd não pode trancar o dono do lado de fora.
func splitEmails(raw string) []string {
	var emails []string
	for _, item := range strings.Split(raw, ",") {
		email := strings.ToLower(strings.TrimSpace(item))
		if email == "" {
			continue
		}
		emails = append(emails, email)
	}

	return emails
}

// loadStorage lê o R2: as quatro variáveis juntas, ou nenhuma. Faltar só uma é erro de
// digitação no systemd, e subir sem arquivos em silêncio esconderia o erro até o primeiro
// upload.
func loadStorage(getenv func(string) string) (Storage, error) {
	storage := Storage{
		Endpoint:        strings.TrimSpace(getenv("R2_ENDPOINT")),
		AccessKeyID:     strings.TrimSpace(getenv("R2_ACCESS_KEY_ID")),
		SecretAccessKey: strings.TrimSpace(getenv("R2_SECRET_ACCESS_KEY")),
		Bucket:          strings.TrimSpace(getenv("R2_BUCKET")),
	}

	filled := 0
	for _, value := range []string{storage.Endpoint, storage.AccessKeyID, storage.SecretAccessKey, storage.Bucket} {
		if value != "" {
			filled++
		}
	}
	if filled != 0 && filled != 4 {
		return Storage{}, errors.New("R2_ENDPOINT, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY and R2_BUCKET must be set together")
	}

	return storage, nil
}

// IsTest diz se o processo está no modo do e2e.
func (cfg Config) IsTest() bool {
	return cfg.Env == EnvTest
}

// loadEnv lê o APP_ENV. O modo de teste com o Neon Auth configurado é recusado: é sinal de
// que o APP_ENV=test escapou para um ambiente de verdade, e aí qualquer pessoa entraria
// como o usuário de teste.
func loadEnv(getenv func(string) string) (string, error) {
	env := strings.TrimSpace(getenv("APP_ENV"))
	if env == "" {
		return EnvDevelopment, nil
	}
	if env != EnvDevelopment && env != EnvTest && env != EnvProduction {
		return "", fmt.Errorf("APP_ENV must be development, test or production, got %q", env)
	}
	if env == EnvTest && strings.TrimSpace(getenv("NEON_AUTH_JWKS_URL")) != "" {
		return "", errors.New("APP_ENV=test never talks to Neon Auth: unset NEON_AUTH_JWKS_URL")
	}

	return env, nil
}
