// Package artifacts é a central de documentos e artefatos (briefing, módulos 1 a 3): a
// galeria, o detalhe com as versões e o que muda sem subir versão nova.
//
// Os tipos daqui saem direto como JSON para a tela, com os nomes de campo de
// svelte/src/lib/contracts/artifact.schema.ts — mudou lá, muda aqui. Um tipo só para banco e
// resposta evita a cópia campo a campo, que é onde os dois lados costumam deixar de combinar.
package artifacts

import (
	"errors"
	"time"
)

// ErrConflict é outra publicação da mesma label terminando primeiro.
var ErrConflict = errors.New("artifacts: concurrent publish")

// thumbFile é o nome da miniatura dentro da pasta da versão, no R2 e no link servido. Começa
// com ponto porque nenhum arquivo de .zip publicado chega com esse nome (os ocultos são
// descartados): a miniatura nunca sobrescreve um arquivo do próprio artefato.
const thumbFile = ".thumb.png"

// Os enums do artifact.schema.ts (e dos CHECK da migration 0001).
var (
	knownKinds    = map[string]bool{"document": true, "plan": true, "idea": true}
	knownStatuses = map[string]bool{"draft": true, "active": true, "done": true, "archived": true}
)

// Artifact é o artifactSchema.
type Artifact struct {
	ID             string     `json:"id"`
	Label          string     `json:"label"`
	Title          string     `json:"title"`
	Description    *string    `json:"description"`
	Kind           string     `json:"kind"`
	Format         string     `json:"format"`
	Status         *string    `json:"status"`
	IsPrivate      bool       `json:"isPrivate"`
	IsPinned       bool       `json:"isPinned"`
	URL            string     `json:"url"`
	ThumbURL       *string    `json:"thumbUrl"`
	CurrentVersion int        `json:"currentVersion"`
	SizeBytes      int64      `json:"sizeBytes"`
	Tags           []string   `json:"tags"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	LastViewedAt   *time.Time `json:"lastViewedAt"`

	// linkToken e hasThumb vêm do banco e viram URL e ThumbURL no Linker; não vão no JSON.
	linkToken string
	hasThumb  bool
}

// Version é o artifactVersionSchema.
type Version struct {
	Version   int       `json:"version"`
	SizeBytes int64     `json:"sizeBytes"`
	HasThumb  bool      `json:"hasThumb"`
	CreatedAt time.Time `json:"createdAt"`
}

// Detail é o artifactDetailSchema: o artefato e todas as versões, da mais nova à mais
// antiga. O Artifact embutido sai "achatado" no JSON, como o .extend() do Zod.
type Detail struct {
	Artifact
	Versions []Version `json:"versions"`
}

// Filter é a query string de GET /api/artifacts. Campo vazio não filtra.
type Filter struct {
	Kind   string
	Search string
}

// Patch é o artifactPatchSchema: só muda o que veio. O status distingue "não veio"
// (StatusSet false) de "veio null" (StatusSet true, Status nil), que limpa a situação.
type Patch struct {
	IsPinned  *bool
	IsPrivate *bool
	StatusSet bool
	Status    *string
}

// Served é o que o link /a/<token>/ precisa para servir: qual pasta do R2 e se exige
// assinatura.
type Served struct {
	ArtifactID string
	Version    int
	Format     string
	IsPrivate  bool
}

// PublishTarget é onde a próxima versão cai: o artefato existente e a maior versão dele.
type PublishTarget struct {
	Exists        bool
	ArtifactID    string
	LinkToken     string
	LatestVersion int
}

// PublishMeta são os campos que acompanham o arquivo. Nil é "não veio": num artefato que
// já existe, mantém o valor; num novo, título e kind são obrigatórios.
type PublishMeta struct {
	Title       *string
	Kind        *string
	Description *string
	Tags        []string
}

// VersionRecord é a versão já enviada ao R2, pronta para entrar no banco.
type VersionRecord struct {
	IsNew       bool
	ArtifactID  string
	OwnerID     string
	Label       string
	Meta        PublishMeta
	Format      string
	ContentText string
	Version     int
	SizeBytes   int64
	HasThumb    bool
}

// Deleted é o que sobra de um artefato apagado: onde estão os arquivos e qual link esquecer.
type Deleted struct {
	ArtifactID string
	LinkToken  string
}
