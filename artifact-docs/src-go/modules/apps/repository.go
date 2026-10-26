package apps

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
)

var (
	//go:embed queries/list_app_types.sql
	listTypesSQL string
	//go:embed queries/list_apps.sql
	listAppsSQL string
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
		if err := rows.Scan(&link.ID, &link.TypeID, &link.Name, &link.URL, &link.Status, &link.CheckedAt); err != nil {
			return nil, err
		}
		link.CheckedAt = database.UTCOrNil(link.CheckedAt)
		links = append(links, link)
	}

	return links, rows.Err()
}
