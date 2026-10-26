// Package store é todo o SQL do hub: artefatos, apps e a pergunta "o dono já tem conta?".
//
// Os tipos daqui saem direto como JSON para a tela, com os nomes de campo dos schemas Zod de
// svelte/src/lib/contracts/ — mudou lá, muda aqui. Um tipo só para banco e resposta evita a
// cópia campo a campo, que é onde os dois lados costumam deixar de combinar.
package store

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

// ErrNotFound é o registro que não existe — ou que existe, mas é de outra pessoa. As duas
// coisas respondem igual, para não contar a ninguém o que há na conta dos outros.
var ErrNotFound = errors.New("store: not found")

// Store fala com o Postgres do Neon.
type Store struct {
	pool *pgxpool.Pool
}

// Open abre o pool de conexões. Não conecta ainda: a primeira conexão nasce na primeira
// consulta (no Neon free, o compute pode estar suspenso e levar ~1 s para acordar).
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("store: parse DATABASE_URL: %w", err)
	}
	config.MaxConns = maxConnections
	// Conexão parada é fechada logo: o Neon suspende o compute ocioso, e a conexão antiga
	// voltaria quebrada.
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("store: open pool: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Begin abre uma transação. É o que o aplicador de migrations usa (migrations.Database).
func (store *Store) Begin(ctx context.Context) (pgx.Tx, error) {
	return store.pool.Begin(ctx)
}

// Close fecha as conexões quando o processo para.
func (store *Store) Close() {
	store.pool.Close()
}

// utc deixa a hora no formato que o z.string().datetime() da tela aceita: terminada em Z. O
// pgx devolve no fuso da máquina, e "-03:00" seria recusado pelo schema.
func utc(moment time.Time) time.Time {
	return moment.UTC()
}

func utcOrNil(moment *time.Time) *time.Time {
	if moment == nil {
		return nil
	}
	converted := moment.UTC()

	return &converted
}

func notFoundOr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	return err
}
