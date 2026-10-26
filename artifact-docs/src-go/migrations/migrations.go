// Package migrations guarda o SQL do banco e o aplica quando o processo sobe.
//
// Os arquivos <NNNN>_<nome>.sql ficam embutidos no binário, e cada um roda uma vez só, em
// ordem, dentro de uma transação. A tabela schema_migrations registra o que já rodou.
//
// É um aplicador nosso, de poucas linhas, em vez de goose ou golang-migrate: para um banco
// de uma pessoa, "rode em ordem o que ainda não rodou" é tudo o que se precisa, e cada
// dependência a menos é binário e memória a menos na VPS de 1 GB.
package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"

	"github.com/jackc/pgx/v5"
)

//go:embed *.sql
var files embed.FS

// namePattern é o formato obrigatório do arquivo: 0001_artifacts.sql. O número é a ordem.
var namePattern = regexp.MustCompile(`^\d{4}_[a-z0-9_]+\.sql$`)

// Migration é um arquivo de SQL, identificado pelo nome.
type Migration struct {
	Name string
	SQL  string
}

// Database é o que o aplicador precisa do banco — o pgxpool.Pool em produção.
type Database interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// lockKey é a chave da trava do Postgres que impede dois processos de aplicarem a mesma
// migration ao mesmo tempo (ex.: o deploy novo subindo antes de o antigo cair). O número é
// arbitrário; só precisa ser sempre o mesmo.
const lockKey = 7_301_202_601

const createTable = `create table if not exists schema_migrations (
	name       text primary key,
	applied_at timestamptz not null default now()
)`

// List devolve as migrations embutidas, em ordem.
//
// Arquivo fora do padrão é erro, e não é ignorado: um 001_x.sql (três dígitos) passaria
// despercebido e a tabela nunca seria criada.
func List(source fs.FS) ([]Migration, error) {
	names, err := fs.Glob(source, "*.sql")
	if err != nil {
		return nil, fmt.Errorf("migrations: list: %w", err)
	}
	sort.Strings(names)

	migrations := make([]Migration, 0, len(names))
	for _, name := range names {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("migrations: %q must be named like 0001_name.sql", name)
		}

		content, err := fs.ReadFile(source, name)
		if err != nil {
			return nil, fmt.Errorf("migrations: read %s: %w", name, err)
		}
		migrations = append(migrations, Migration{Name: name, SQL: string(content)})
	}

	return migrations, nil
}

// Embedded são as migrations que vieram dentro do binário.
func Embedded() ([]Migration, error) {
	return List(files)
}

// Apply roda, em ordem, as migrations que ainda não rodaram, e devolve os nomes aplicados.
//
// Cada uma roda numa transação própria: se a 0003 falhar, a 0001 e a 0002 continuam
// aplicadas, e a 0003 não fica pela metade.
func Apply(ctx context.Context, database Database, migrations []Migration) ([]string, error) {
	var applied []string
	for _, migration := range migrations {
		ran, err := applyOne(ctx, database, migration)
		if err != nil {
			return applied, fmt.Errorf("migrations: %s: %w", migration.Name, err)
		}
		if ran {
			applied = append(applied, migration.Name)
		}
	}

	return applied, nil
}

func applyOne(ctx context.Context, database Database, migration Migration) (bool, error) {
	transaction, err := database.Begin(ctx)
	if err != nil {
		return false, err
	}
	// Depois do Commit, o Rollback não faz nada; antes dele, desfaz tudo.
	defer func() { _ = transaction.Rollback(ctx) }()

	if _, err := transaction.Exec(ctx, "select pg_advisory_xact_lock($1)", lockKey); err != nil {
		return false, err
	}
	if _, err := transaction.Exec(ctx, createTable); err != nil {
		return false, err
	}

	var done bool
	err = transaction.QueryRow(ctx,
		"select exists (select 1 from schema_migrations where name = $1)", migration.Name,
	).Scan(&done)
	if err != nil || done {
		return false, err
	}

	// O protocolo simples aceita vários comandos num texto só — é o formato de um .sql.
	if _, err := transaction.Exec(ctx, migration.SQL, pgx.QueryExecModeSimpleProtocol); err != nil {
		return false, err
	}
	if _, err := transaction.Exec(ctx, "insert into schema_migrations (name) values ($1)", migration.Name); err != nil {
		return false, err
	}

	return true, transaction.Commit(ctx)
}
