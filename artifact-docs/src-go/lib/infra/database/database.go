// Package database abre a conexão com o Postgres do Neon e mantém o schema em dia.
//
// O SQL de cada feature mora no módulo dela (modules/<feature>/queries/); aqui fica só o que
// é do banco como um todo: o pool de conexões e as migrations.
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// maxConnections é pequeno de propósito (briefing, "Detalhes do Go"): o endpoint pooled do
// Neon já multiplexa as conexões, e cada conexão aberta é memória na VPS de 1 GB.
const maxConnections = 4

// maxIdleTime fecha logo a conexão parada: o Neon suspende o compute ocioso, e a conexão
// antiga voltaria quebrada.
const maxIdleTime = 5 * time.Minute

// ErrNotFound é o registro que não existe — ou que existe, mas é de outra pessoa. As duas
// coisas respondem igual, para não contar a ninguém o que há na conta dos outros.
var ErrNotFound = errors.New("database: not found")

// Open abre o pool de conexões. Não conecta ainda: a primeira conexão nasce na primeira
// consulta (no Neon free, o compute pode estar suspenso e levar ~1 s para acordar).
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("database: parse DATABASE_URL: %w", err)
	}
	config.MaxConns = maxConnections
	config.MaxConnIdleTime = maxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("database: open pool: %w", err)
	}

	return pool, nil
}

// NotFoundOr troca o "nenhuma linha" do pgx pelo ErrNotFound, que é o que os módulos
// conhecem; qualquer outro erro passa como veio.
func NotFoundOr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	return err
}

// UTC deixa a hora no formato que o z.string().datetime() da tela aceita: terminada em Z. O
// pgx devolve no fuso da máquina, e "-03:00" seria recusado pelo schema.
func UTC(moment time.Time) time.Time {
	return moment.UTC()
}

// UTCOrNil é o UTC para coluna que aceita NULL.
func UTCOrNil(moment *time.Time) *time.Time {
	if moment == nil {
		return nil
	}
	converted := moment.UTC()

	return &converted
}
