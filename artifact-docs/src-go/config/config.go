// Package config lê a configuração do processo a partir do ambiente.
//
// Só variável de ambiente, nada de arquivo: em desenvolvimento quem as carrega é o mise (a
// partir de src-go/.env); em produção, o systemd. O código não sabe de onde elas vieram.
package config

import (
	"fmt"
	"strconv"
	"strings"
)

// DefaultPort é a porta da API em desenvolvimento — a mesma para a qual o Vite encaminha /api
// (VITE_API_PORT em svelte/.env.example).
const DefaultPort = 3010

// Config é tudo o que o processo precisa saber para subir.
type Config struct {
	Port int
	// NeonAuthURL é o endereço do login (Auth → URL do endpoint, no painel da Neon).
	NeonAuthURL string
	// NeonAuthJWKSURL é de onde vêm as chaves que conferem o token. Vazia, o login fica
	// desligado e as rotas protegidas respondem 503 — em desenvolvimento a tela usa o mock.
	NeonAuthJWKSURL string
	// AllowedEmails é quem pode usar o hub, em minúsculas (AUTH_ALLOWED_EMAILS, separados
	// por vírgula).
	AllowedEmails []string
}

// Addr é o endereço em que o servidor escuta.
func (c Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
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

	return Config{
		Port:            port,
		NeonAuthURL:     strings.TrimSpace(getenv("NEON_AUTH_URL")),
		NeonAuthJWKSURL: strings.TrimSpace(getenv("NEON_AUTH_JWKS_URL")),
		AllowedEmails:   splitEmails(getenv("AUTH_ALLOWED_EMAILS")),
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
