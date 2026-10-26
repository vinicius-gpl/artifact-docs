// Package apps é o hub de apps (briefing, módulo 4): os apps cadastrados, agrupados pelos
// tipos que a pessoa cria.
//
// Os tipos daqui saem direto como JSON para a tela, com os nomes de campo de
// svelte/src/lib/contracts/app-link.schema.ts — mudou lá, muda aqui.
package apps

import "time"

// Type é o appTypeSchema.
type Type struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

// Link é o appLinkSchema. Chama Link, e não App, pelo mesmo motivo do contrato: "app" já é o
// nome do próprio sistema em todo canto.
type Link struct {
	ID     string `json:"id"`
	TypeID string `json:"typeId"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	// LogoURL fica null até existir a rota que serve arquivo do R2; a tela usa o favicon.
	LogoURL   *string    `json:"logoUrl"`
	Status    string     `json:"status"`
	CheckedAt *time.Time `json:"checkedAt"`
}

// Hub é o appHubSchema: tipos e apps juntos, porque a tela sempre agrupa.
type Hub struct {
	Types []Type `json:"types"`
	Apps  []Link `json:"apps"`
}
