package apps

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
)

var (
	//go:embed queries/list_app_types.sql
	listTypesSQL string
	//go:embed queries/list_apps.sql
	listAppsSQL string
	//go:embed queries/upsert_app_type.sql
	upsertTypeSQL string
	//go:embed queries/insert_app.sql
	insertAppSQL string
	//go:embed queries/update_app.sql
	updateAppSQL string
	//go:embed queries/delete_app.sql
	deleteAppSQL string
	//go:embed queries/delete_empty_app_types.sql
	deleteEmptyTypesSQL string
	//go:embed queries/list_app_targets.sql
	listTargetsSQL string
	//go:embed queries/update_app_status.sql
	updateStatusSQL string
	//go:embed queries/update_app_icon.sql
	updateIconSQL string
)

// Repository lê o hub de apps no Postgres.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository lê do pool do Neon.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Hub devolve os tipos e os apps.
func (repository *Repository) Hub(ctx context.Context) (Hub, error) {
	types, err := repository.listTypes(ctx)
	if err != nil {
		return Hub{}, err
	}

	links, err := repository.listLinks(ctx)
	if err != nil {
		return Hub{}, err
	}

	return Hub{Types: types, Apps: links}, nil
}

func (repository *Repository) listTypes(ctx context.Context) ([]Type, error) {
	rows, err := repository.pool.Query(ctx, listTypesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := []Type{}
	for rows.Next() {
		var appType Type
		if err := rows.Scan(&appType.ID, &appType.Name, &appType.Position); err != nil {
			return nil, err
		}
		types = append(types, appType)
	}

	return types, rows.Err()
}

func (repository *Repository) listLinks(ctx context.Context) ([]Link, error) {
	rows, err := repository.pool.Query(ctx, listAppsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []Link{}
	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.ID, &link.TypeID, &link.Name, &link.URL, &link.LogoURL, &link.Status, &link.CheckedAt); err != nil {
			return nil, err
		}
		link.CheckedAt = database.UTCOrNil(link.CheckedAt)
		links = append(links, link)
	}

	return links, rows.Err()
}

// Create cadastra o app, criando o tipo se ele ainda não existe — numa transação: um tipo
// novo nunca fica para trás sem o app que o pediu.
func (repository *Repository) Create(ctx context.Context, input Input) (Link, error) {
	transaction, err := repository.pool.Begin(ctx)
	if err != nil {
		return Link{}, err
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	typeID, err := upsertType(ctx, transaction, input.TypeName)
	if err != nil {
		return Link{}, err
	}
	link, err := scanLink(transaction.QueryRow(ctx, insertAppSQL, typeID, input.Name, input.URL))
	if err != nil {
		return Link{}, err
	}

	return link, transaction.Commit(ctx)
}

// Update edita o app. O tipo que ficou vazio (o app mudou de tipo) some na mesma transação.
func (repository *Repository) Update(ctx context.Context, id string, input Input) (Link, error) {
	transaction, err := repository.pool.Begin(ctx)
	if err != nil {
		return Link{}, err
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	typeID, err := upsertType(ctx, transaction, input.TypeName)
	if err != nil {
		return Link{}, err
	}
	link, err := scanLink(transaction.QueryRow(ctx, updateAppSQL, id, typeID, input.Name, input.URL))
	if err != nil {
		return Link{}, database.NotFoundOr(err)
	}
	if _, err := transaction.Exec(ctx, deleteEmptyTypesSQL); err != nil {
		return Link{}, err
	}

	return link, transaction.Commit(ctx)
}

// Delete apaga o app e, se ele era o último do tipo, o tipo junto.
func (repository *Repository) Delete(ctx context.Context, id string) error {
	transaction, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	deleted, err := transaction.Exec(ctx, deleteAppSQL, id)
	if err != nil {
		return err
	}
	if deleted.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	if _, err := transaction.Exec(ctx, deleteEmptyTypesSQL); err != nil {
		return err
	}

	return transaction.Commit(ctx)
}

// Targets são todos os apps, para o monitor conferir.
func (repository *Repository) Targets(ctx context.Context) ([]Target, error) {
	rows, err := repository.pool.Query(ctx, listTargetsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targets := []Target{}
	for rows.Next() {
		var target Target
		if err := rows.Scan(&target.ID, &target.URL, &target.IconURL); err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	return targets, rows.Err()
}

// SetStatus grava o resultado de uma conferência.
func (repository *Repository) SetStatus(ctx context.Context, id, status string) error {
	_, err := repository.pool.Exec(ctx, updateStatusSQL, id, status)

	return err
}

// SetIcon grava o ícone que o monitor achou.
func (repository *Repository) SetIcon(ctx context.Context, id, iconURL string) error {
	_, err := repository.pool.Exec(ctx, updateIconSQL, id, iconURL)

	return err
}

func upsertType(ctx context.Context, transaction pgx.Tx, name string) (string, error) {
	var typeID string
	err := transaction.QueryRow(ctx, upsertTypeSQL, name).Scan(&typeID)

	return typeID, err
}

// scanLink lê uma linha de insert_app.sql ou update_app.sql — as mesmas colunas, na mesma
// ordem.
func scanLink(row pgx.Row) (Link, error) {
	var link Link
	err := row.Scan(&link.ID, &link.TypeID, &link.Name, &link.URL, &link.LogoURL, &link.Status, &link.CheckedAt)
	link.CheckedAt = database.UTCOrNil(link.CheckedAt)

	return link, err
}
