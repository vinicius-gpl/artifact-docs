package artifacts

import (
	"context"
	_ "embed"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	//go:embed queries/update_artifact.sql
	updateArtifactSQL string
	//go:embed queries/find_served_artifact.sql
	findServedSQL string
	//go:embed queries/find_publish_target.sql
	findPublishTargetSQL string
	//go:embed queries/insert_artifact.sql
	insertArtifactSQL string
	//go:embed queries/update_published_artifact.sql
	updatePublishedSQL string
	//go:embed queries/insert_artifact_version.sql
	insertVersionSQL string
	//go:embed queries/restore_artifact_version.sql
	restoreVersionSQL string
	//go:embed queries/delete_artifact.sql
	deleteArtifactSQL string
)

// uniqueViolation é o código do Postgres para chave repetida: duas publicações da mesma
// label ao mesmo tempo.
const uniqueViolation = "23505"

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

	artifact, err := repository.Find(ctx, ownerID, label)
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
	updated, err := repository.pool.Exec(ctx, updateArtifactSQL,
		ownerID, label, patch.IsPinned, patch.IsPrivate, patch.StatusSet, patch.Status,
		patch.Title, patch.DescriptionSet, patch.Description, patch.Tags)
	if err != nil {
		return Artifact{}, err
	}
	if updated.RowsAffected() == 0 {
		return Artifact{}, database.ErrNotFound
	}

	return repository.Find(ctx, ownerID, label)
}

// Find devolve um artefato do dono, sem marcar acesso.
func (repository *Repository) Find(ctx context.Context, ownerID, label string) (Artifact, error) {
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
	err := row.Scan(
		&artifact.ID, &artifact.Label, &artifact.Title, &artifact.Description, &artifact.Kind,
		&artifact.Format, &artifact.Status, &artifact.linkToken, &artifact.IsPrivate, &artifact.IsPinned,
		&artifact.CurrentVersion, &artifact.SizeBytes, &artifact.hasThumb, &artifact.Tags,
		&artifact.CreatedAt, &artifact.UpdatedAt, &artifact.LastViewedAt,
	)
	if err != nil {
		return Artifact{}, err
	}

	return normalized(artifact), nil
}

// normalized deixa o artefato no formato que o JSON da tela aceita: lista vazia em vez de
// nula, e hora em UTC.
func normalized(artifact Artifact) Artifact {
	if artifact.Tags == nil {
		artifact.Tags = []string{}
	}
	artifact.CreatedAt = database.UTC(artifact.CreatedAt)
	artifact.UpdatedAt = database.UTC(artifact.UpdatedAt)
	artifact.LastViewedAt = database.UTCOrNil(artifact.LastViewedAt)

	return artifact
}

// FindServed acha o artefato por trás de um link /a/<token>/.
func (repository *Repository) FindServed(ctx context.Context, linkToken string) (Served, error) {
	var served Served
	err := repository.pool.QueryRow(ctx, findServedSQL, linkToken).
		Scan(
			&served.ArtifactID,
			&served.Version,
			&served.Format,
			&served.IsPrivate,
			&served.Title,
			&served.Description,
			&served.Kind,
			&served.HasThumb,
		)

	return served, database.NotFoundOr(err)
}

// PublishTarget diz onde a próxima versão vai cair. Exists false é artefato novo.
func (repository *Repository) PublishTarget(ctx context.Context, ownerID, label string) (PublishTarget, error) {
	target := PublishTarget{Exists: true}
	err := repository.pool.QueryRow(ctx, findPublishTargetSQL, ownerID, label).
		Scan(&target.ArtifactID, &target.LinkToken, &target.LatestVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublishTarget{}, nil
	}

	return target, err
}

// SaveVersion grava a versão já enviada ao R2 e troca o ponteiro, numa transação: ou a
// versão nova aparece inteira, ou nada muda.
func (repository *Repository) SaveVersion(ctx context.Context, record VersionRecord) error {
	transaction, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	if err := saveArtifactRow(ctx, transaction, record); err != nil {
		return conflictOr(err)
	}
	_, err = transaction.Exec(ctx, insertVersionSQL, record.ArtifactID, record.Version, record.SizeBytes, record.HasThumb)
	if err != nil {
		return conflictOr(err)
	}

	return transaction.Commit(ctx)
}

func saveArtifactRow(ctx context.Context, transaction pgx.Tx, record VersionRecord) error {
	meta := record.Meta
	if record.IsNew {
		_, err := transaction.Exec(ctx, insertArtifactSQL,
			record.ArtifactID, record.OwnerID, record.Label, meta.Title, meta.Description,
			meta.Kind, record.Format, meta.Tags, record.ContentText, record.Version,
			meta.Status, meta.IsPrivate)
		return err
	}

	_, err := transaction.Exec(ctx, updatePublishedSQL,
		record.ArtifactID, meta.Title, meta.Description != nil, meta.Description,
		meta.Kind, meta.Tags, record.Format, record.ContentText, record.Version,
		meta.Status, meta.IsPrivate)

	return err
}

// Restore volta o ponteiro para uma versão que existe e devolve o link_token, para o cache
// esquecer a versão antiga.
func (repository *Repository) Restore(ctx context.Context, ownerID, label string, version int) (string, error) {
	var linkToken string
	err := repository.pool.QueryRow(ctx, restoreVersionSQL, ownerID, label, version).Scan(&linkToken)

	return linkToken, database.NotFoundOr(err)
}

// Delete apaga o artefato (as versões vão junto) e devolve o id e o link_token.
func (repository *Repository) Delete(ctx context.Context, ownerID, label string) (Deleted, error) {
	var deleted Deleted
	err := repository.pool.QueryRow(ctx, deleteArtifactSQL, ownerID, label).
		Scan(&deleted.ArtifactID, &deleted.LinkToken)

	return deleted, database.NotFoundOr(err)
}

func conflictOr(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == uniqueViolation {
		return ErrConflict
	}

	return err
}
