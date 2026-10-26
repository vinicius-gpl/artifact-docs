package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Artifact é o artifactSchema de svelte/src/lib/contracts/artifact.schema.ts.
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
}

// ArtifactVersion é o artifactVersionSchema.
type ArtifactVersion struct {
	Version   int       `json:"version"`
	SizeBytes int64     `json:"sizeBytes"`
	HasThumb  bool      `json:"hasThumb"`
	CreatedAt time.Time `json:"createdAt"`
}

// ArtifactDetail é o artifactDetailSchema: o artefato e todas as versões, da mais nova à
// mais antiga. O Artifact embutido sai "achatado" no JSON, como o .extend() do Zod.
type ArtifactDetail struct {
	Artifact
	Versions []ArtifactVersion `json:"versions"`
}

// ArtifactFilter é a query string de GET /api/artifacts. Campo vazio não filtra.
type ArtifactFilter struct {
	Kind   string
	Search string
}

// ArtifactPatch é o artifactPatchSchema: só muda o que veio. O status distingue "não veio"
// (StatusSet false) de "veio null" (StatusSet true, Status nil), que limpa a situação.
type ArtifactPatch struct {
	IsPinned  *bool
	IsPrivate *bool
	StatusSet bool
	Status    *string
}

// artifactColumns lê o artefato de `source` com o tamanho e a miniatura da versão servida.
// `source` é a tabela ou um CTE (o resultado de um update ... returning).
func artifactColumns(source string) string {
	return fmt.Sprintf(`select artifact.id::text, artifact.label, artifact.title,
	artifact.description, artifact.kind, artifact.format, artifact.status, artifact.link_token,
	artifact.is_private, artifact.is_pinned, artifact.current_version,
	coalesce(served.size_bytes, 0), coalesce(served.has_thumb, false), artifact.tags,
	artifact.created_at, artifact.updated_at, artifact.last_viewed_at
from %s artifact
left join artifact_versions served
	on served.artifact_id = artifact.id and served.version = artifact.current_version`, source)
}

func scanArtifact(row pgx.Row) (Artifact, error) {
	var artifact Artifact
	var linkToken string
	var hasThumb bool

	err := row.Scan(
		&artifact.ID, &artifact.Label, &artifact.Title, &artifact.Description, &artifact.Kind,
		&artifact.Format, &artifact.Status, &linkToken, &artifact.IsPrivate, &artifact.IsPinned,
		&artifact.CurrentVersion, &artifact.SizeBytes, &hasThumb, &artifact.Tags,
		&artifact.CreatedAt, &artifact.UpdatedAt, &artifact.LastViewedAt,
	)
	if err != nil {
		return Artifact{}, err
	}

	return withLinks(artifact, linkToken, hasThumb), nil
}

// withLinks monta os endereços servidos e normaliza o que o JSON da tela não aceita (lista
// nula, hora fora de UTC). A tela nunca monta o link sozinha: ela recebe pronto.
func withLinks(artifact Artifact, linkToken string, hasThumb bool) Artifact {
	artifact.URL = "/a/" + linkToken + "/"
	if hasThumb {
		thumb := artifact.URL + "thumb.png"
		artifact.ThumbURL = &thumb
	}
	if artifact.Tags == nil {
		artifact.Tags = []string{}
	}
	artifact.CreatedAt = utc(artifact.CreatedAt)
	artifact.UpdatedAt = utc(artifact.UpdatedAt)
	artifact.LastViewedAt = utcOrNil(artifact.LastViewedAt)

	return artifact
}

// ListArtifacts devolve os artefatos do dono, do editado mais recente ao mais antigo.
//
// A busca usa websearch_to_tsquery: aceita o que a pessoa digita ("cnpj -abertura", frases
// entre aspas) sem nunca dar erro de sintaxe.
func (store *Store) ListArtifacts(ctx context.Context, ownerID string, filter ArtifactFilter) ([]Artifact, error) {
	query := artifactColumns("artifacts") + `
where artifact.owner_id = $1
	and ($2 = '' or artifact.kind = $2)
	and ($3 = '' or artifact.search @@ websearch_to_tsquery('portuguese', $3))
order by artifact.updated_at desc`

	rows, err := store.pool.Query(ctx, query, ownerID, filter.Kind, filter.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artifacts := []Artifact{}
	for rows.Next() {
		artifact, err := scanArtifact(rows)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, rows.Err()
}

// OpenArtifact devolve o artefato com as versões e marca o acesso (last_viewed_at), que é o
// que ordena "Recentes" na galeria.
func (store *Store) OpenArtifact(ctx context.Context, ownerID, label string) (ArtifactDetail, error) {
	query := `with opened as (
	update artifacts set last_viewed_at = now()
	where owner_id = $1 and label = $2
	returning *
)
` + artifactColumns("opened")

	artifact, err := scanArtifact(store.pool.QueryRow(ctx, query, ownerID, label))
	if err != nil {
		return ArtifactDetail{}, notFoundOr(err)
	}

	versions, err := store.listVersions(ctx, artifact.ID)
	if err != nil {
		return ArtifactDetail{}, err
	}

	return ArtifactDetail{Artifact: artifact, Versions: versions}, nil
}

func (store *Store) listVersions(ctx context.Context, artifactID string) ([]ArtifactVersion, error) {
	rows, err := store.pool.Query(ctx, `select version, size_bytes, has_thumb, created_at
from artifact_versions where artifact_id = $1 order by version desc`, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := []ArtifactVersion{}
	for rows.Next() {
		var version ArtifactVersion
		if err := rows.Scan(&version.Version, &version.SizeBytes, &version.HasThumb, &version.CreatedAt); err != nil {
			return nil, err
		}
		version.CreatedAt = utc(version.CreatedAt)
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// PatchArtifact muda o que não pede versão nova: fixar, privado e situação.
//
// Não mexe em updated_at: fixar um card não é editar o artefato, e o "Editado há X" da
// galeria precisa continuar dizendo quando o conteúdo mudou.
func (store *Store) PatchArtifact(ctx context.Context, ownerID, label string, patch ArtifactPatch) (Artifact, error) {
	query := `with patched as (
	update artifacts set
		is_pinned  = coalesce($3::boolean, is_pinned),
		is_private = coalesce($4::boolean, is_private),
		status     = case when $5::boolean then $6::text else status end
	where owner_id = $1 and label = $2
	returning *
)
` + artifactColumns("patched")

	row := store.pool.QueryRow(ctx, query,
		ownerID, label, patch.IsPinned, patch.IsPrivate, patch.StatusSet, patch.Status)
	artifact, err := scanArtifact(row)
	if err != nil {
		return Artifact{}, notFoundOr(err)
	}

	return artifact, nil
}
