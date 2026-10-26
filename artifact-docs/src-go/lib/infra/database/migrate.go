package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"

	"github.com/jackc/pgx/v5"
)

// As migrations ficam agrupadas por assunto, como no acerola-reader:
//
//	migrations/models/<domínio>/NNNN_create_<tabela>.sql   tabelas, gatilhos e índices
//	migrations/views/NNNN_create_<nome>_view.sql            views
//
// A pasta só organiza; quem manda na ordem é o número, que é único no projeto inteiro. Uma
// migration nova pega o próximo número, seja qual for a pasta.
//
//go:embed migrations
var embedded embed.FS

// namePattern é o formato obrigatório do arquivo: 0001_create_artifacts.sql.
var namePattern = regexp.MustCompile(`^(\d{4})_[a-z0-9_]+\.sql$`)

// lockKey é a chave da trava do Postgres que impede dois processos de aplicarem a mesma
// migration ao mesmo tempo (ex.: o deploy novo subindo antes de o antigo cair). O número é
// arbitrário; só precisa ser sempre o mesmo.
const lockKey = 7_301_202_601

const (
	createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
  name       TEXT        PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`
	lockMigrations   = `SELECT pg_advisory_xact_lock($1)`
	migrationApplied = `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name = $1)`
	recordMigration  = `INSERT INTO schema_migrations (name) VALUES ($1)`
)

// Migration é um arquivo de SQL. Name é o nome do arquivo, sem a pasta: é ele que fica
// registrado, para que mover um arquivo de pasta não o faça rodar de novo.
type Migration struct {
	Name string
	SQL  string
}

// Transactor é o que o aplicador precisa do banco — o *pgxpool.Pool em produção.
type Transactor interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Embedded são as migrations que vieram dentro do binário, em ordem.
func Embedded() ([]Migration, error) {
	return List(embedded)
}

// List lê todo .sql de source, em qualquer subpasta, e devolve em ordem de número.
//
// Arquivo fora do padrão ou número repetido é erro, e não é ignorado: um 001_x.sql (três
// dígitos) passaria despercebido, e dois 0007 rodariam numa ordem que ninguém escolheu.
func List(source fs.FS) ([]Migration, error) {
	var migrations []Migration
	err := fs.WalkDir(source, ".", func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}

		migration, err := readMigration(source, filePath)
		if err != nil {
			return err
		}
		migrations = append(migrations, migration)

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(migrations, func(left, right int) bool {
		return migrations[left].Name < migrations[right].Name
	})

	return migrations, checkUniqueNumbers(migrations)
}

func readMigration(source fs.FS, filePath string) (Migration, error) {
	name := path.Base(filePath)
	if !namePattern.MatchString(name) {
		return Migration{}, fmt.Errorf("database: %q must be named like 0001_create_table.sql", filePath)
	}

	content, err := fs.ReadFile(source, filePath)
	if err != nil {
		return Migration{}, fmt.Errorf("database: read %s: %w", filePath, err)
	}

	return Migration{Name: name, SQL: string(content)}, nil
}

func checkUniqueNumbers(migrations []Migration) error {
	seen := make(map[string]string, len(migrations))
	for _, migration := range migrations {
		number := namePattern.FindStringSubmatch(migration.Name)[1]
		if previous, taken := seen[number]; taken {
			return fmt.Errorf("database: %s and %s share the number %s", previous, migration.Name, number)
		}
		seen[number] = migration.Name
	}

	return nil
}

// Migrate roda, em ordem, as migrations que ainda não rodaram, e devolve os nomes aplicados.
//
// Cada uma roda numa transação própria: se a 0003 falhar, a 0001 e a 0002 continuam
// aplicadas, e a 0003 não fica pela metade.
func Migrate(ctx context.Context, database Transactor, migrations []Migration) ([]string, error) {
	var applied []string
	for _, migration := range migrations {
		ran, err := applyOne(ctx, database, migration)
		if err != nil {
			return applied, fmt.Errorf("database: migration %s: %w", migration.Name, err)
		}
		if ran {
			applied = append(applied, migration.Name)
		}
	}

	return applied, nil
}

func applyOne(ctx context.Context, database Transactor, migration Migration) (bool, error) {
	transaction, err := database.Begin(ctx)
	if err != nil {
		return false, err
	}
	// Depois do Commit, o Rollback não faz nada; antes dele, desfaz tudo.
	defer func() { _ = transaction.Rollback(ctx) }()

	if _, err := transaction.Exec(ctx, lockMigrations, lockKey); err != nil {
		return false, err
	}
	if _, err := transaction.Exec(ctx, createMigrationsTable); err != nil {
		return false, err
	}

	var done bool
	if err := transaction.QueryRow(ctx, migrationApplied, migration.Name).Scan(&done); err != nil || done {
		return false, err
	}

	// O protocolo simples aceita vários comandos num texto só — é o formato de um .sql.
	if _, err := transaction.Exec(ctx, migration.SQL, pgx.QueryExecModeSimpleProtocol); err != nil {
		return false, err
	}
	if _, err := transaction.Exec(ctx, recordMigration, migration.Name); err != nil {
		return false, err
	}

	return true, transaction.Commit(ctx)
}
