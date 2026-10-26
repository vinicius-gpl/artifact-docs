// Package config lê a configuração do processo a partir do ambiente.
//
// Só variável de ambiente, nada de arquivo: em desenvolvimento quem as carrega é o mise (a
// partir de src-go/.env); em produção, o systemd. O código não sabe de onde elas vieram.
package config

import (
	"fmt"
	"strconv"
)

// DefaultPort é a porta da API em desenvolvimento — a mesma para a qual o Vite encaminha /api
// (VITE_API_PORT em svelte/.env.example).
const DefaultPort = 3010

// Config é tudo o que o processo precisa saber para subir.
type Config struct {
	Port int
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
	cfg := Config{Port: DefaultPort}

	raw := getenv("API_PORT")
	if raw == "" {
		return cfg, nil
	}

	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("API_PORT must be a port between 1 and 65535, got %q", raw)
	}

	cfg.Port = port

	return cfg, nil
}
