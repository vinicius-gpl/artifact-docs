package artifacts

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
)

// O SQL mora em queries/, um arquivo por consulta: fica legível, com destaque de sintaxe, e
// a consulta que roda é exatamente a que está no arquivo — nada é montado com Sprintf.
var (
	//go:embed queries/list_artifacts.sql
	listArtifactsSQL string
	//go:embed queries/find_artifact.sql
	findArtifactSQL string
	//go:embed queries/list_artifact_versions.sql
	listVersionsSQL string
	//go:embed queries/mark_artifact_viewed.sql
	markViewedSQL string
	//go:embed queries/update_artifact_flags.sql
	updateFlagsSQL string
)

// Repository lê e grava os artefatos no Postgres.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository lê e grava no pool do Neon.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// List devolve os artefatos do dono que passam no filtro.
func (repository *Repository) List(ctx context.Context, ownerID string, filter Filter) ([]Artifact, error) {
	rows, err := repository.pool.Query(ctx, listArtifactsSQL, ownerID, filter.Kind, filter.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	found := []Artifact{}
	for rows.Next() {
		artifact, err := scanArtifact(rows)
		if err != nil {
			return nil, err
		}
		found = append(found, artifact)
	}

	return found, rows.Err()
}

// Open marca o acesso e devolve o artefato com as versões.
func (repository *Repository) Open(ctx context.Context, ownerID, label string) (Detail, error) {
	marked, err := repository.pool.Exec(ctx, markViewedSQL, ownerID, label)
	if err != nil {
		return Detail{}, err
	}
	if marked.RowsAffected() == 0 {
		return Detail{}, database.ErrNotFound
	}

	artifact, err := repository.find(ctx, ownerID, label)
	if err != nil {
		return Detail{}, err
	}

	versions, err := repository.listVersions(ctx, artifact.ID)
	if err != nil {
		return Detail{}, err
	}

	return Detail{Artifact: artifact, Versions: versions}, nil
}

// Update aplica o patch e devolve o artefato como ficou.
func (repository *Repository) Update(ctx context.Context, ownerID, label string, patch Patch) (Artifact, error) {
	updated, err := repository.pool.Exec(ctx, updateFlagsSQL,
		ownerID, label, patch.IsPinned, patch.IsPrivate, patch.StatusSet, patch.Status)
	if err != nil {
		return Artifact{}, err
	}
	if updated.RowsAffected() == 0 {
		return Artifact{}, database.ErrNotFound
	}

	return repository.find(ctx, ownerID, label)
}

func (repository *Repository) find(ctx context.Context, ownerID, label string) (Artifact, error) {
	artifact, err := scanArtifact(repository.pool.QueryRow(ctx, findArtifactSQL, ownerID, label))

	return artifact, database.NotFoundOr(err)
}

func (repository *Repository) listVersions(ctx context.Context, artifactID string) ([]Version, error) {
	rows, err := repository.pool.Query(ctx, listVersionsSQL, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := []Version{}
	for rows.Next() {
		var version Version
		if err := rows.Scan(&version.Version, &version.SizeBytes, &version.HasThumb, &version.CreatedAt); err != nil {
			return nil, err
		}
		version.CreatedAt = database.UTC(version.CreatedAt)
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// scanArtifact lê uma linha de list_artifacts.sql ou find_artifact.sql — as duas têm as
// mesmas colunas, na mesma ordem.
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
	artifact.CreatedAt = database.UTC(artifact.CreatedAt)
	artifact.UpdatedAt = database.UTC(artifact.UpdatedAt)
	artifact.LastViewedAt = database.UTCOrNil(artifact.LastViewedAt)

	return artifact
}
