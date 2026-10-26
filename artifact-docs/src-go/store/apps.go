package store

import (
	"context"
	"time"
)

// AppType é o appTypeSchema de svelte/src/lib/contracts/app-link.schema.ts.
type AppType struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

// AppLink é o appLinkSchema.
type AppLink struct {
	ID     string `json:"id"`
	TypeID string `json:"typeId"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	// LogoURL fica null até existir a rota que serve arquivo do R2; a tela usa o favicon.
	LogoURL   *string    `json:"logoUrl"`
	Status    string     `json:"status"`
	CheckedAt *time.Time `json:"checkedAt"`
}

// AppHub é o appHubSchema: tipos e apps juntos, porque a tela sempre agrupa.
type AppHub struct {
	Types []AppType `json:"types"`
	Apps  []AppLink `json:"apps"`
}

// AppHub devolve os tipos na ordem escolhida pela pessoa e os apps em ordem alfabética.
func (store *Store) AppHub(ctx context.Context) (AppHub, error) {
	types, err := store.listAppTypes(ctx)
	if err != nil {
		return AppHub{}, err
	}

	apps, err := store.listApps(ctx)
	if err != nil {
		return AppHub{}, err
	}

	return AppHub{Types: types, Apps: apps}, nil
}

func (store *Store) listAppTypes(ctx context.Context) ([]AppType, error) {
	rows, err := store.pool.Query(ctx,
		"select id::text, name, position from app_types order by position, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := []AppType{}
	for rows.Next() {
		var appType AppType
		if err := rows.Scan(&appType.ID, &appType.Name, &appType.Position); err != nil {
			return nil, err
		}
		types = append(types, appType)
	}

	return types, rows.Err()
}

func (store *Store) listApps(ctx context.Context) ([]AppLink, error) {
	rows, err := store.pool.Query(ctx, `select id::text, type_id::text, name, url, status, checked_at
from apps order by lower(name)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []AppLink{}
	for rows.Next() {
		var app AppLink
		if err := rows.Scan(&app.ID, &app.TypeID, &app.Name, &app.URL, &app.Status, &app.CheckedAt); err != nil {
			return nil, err
		}
		app.CheckedAt = utcOrNil(app.CheckedAt)
		apps = append(apps, app)
	}

	return apps, rows.Err()
}
