package artifacts

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
)

// feliz
func TestWithLinksBuildsTheServedAddressesAndNormalizes(test *testing.T) {
	saoPaulo := time.FixedZone("BRT", -3*60*60)
	viewed := time.Date(2026, 10, 26, 9, 0, 0, 0, saoPaulo)

	artifact := withLinks(Artifact{
		CreatedAt:    time.Date(2026, 10, 1, 9, 0, 0, 0, saoPaulo),
		UpdatedAt:    time.Date(2026, 10, 2, 9, 0, 0, 0, saoPaulo),
		LastViewedAt: &viewed,
	}, "abc123", true)

	if artifact.URL != "/a/abc123/" {
		test.Fatalf("url = %q", artifact.URL)
	}
	if artifact.ThumbURL == nil || *artifact.ThumbURL != "/a/abc123/thumb.png" {
		test.Fatalf("thumbUrl = %v", artifact.ThumbURL)
	}
	if artifact.Tags == nil {
		test.Fatal("tags must be an empty list, never null")
	}
	if artifact.UpdatedAt.Location() != time.UTC || artifact.LastViewedAt.Location() != time.UTC {
		test.Fatal("times must be in UTC, so the JSON ends in Z")
	}
	if artifact.UpdatedAt.Hour() != 12 {
		test.Fatalf("updatedAt hour = %d, want 12 (9h in -03:00)", artifact.UpdatedAt.Hour())
	}
}

// triste
func TestWithLinksLeavesTheThumbEmptyWhenThereIsNone(test *testing.T) {
	artifact := withLinks(Artifact{Tags: []string{"marca"}}, "abc123", false)

	if artifact.ThumbURL != nil {
		test.Fatalf("thumbUrl = %v, want nil", *artifact.ThumbURL)
	}
	if artifact.LastViewedAt != nil {
		test.Fatal("never viewed must stay null")
	}
}

// seededRepository conecta ao banco de teste, aplica as migrations e cria um artefato de um
// dono que só este teste usa — ou pula: o CI não tem banco, e o SQL de verdade só é
// conferido onde a TEST_DATABASE_URL existe (src-go/.env).
func seededRepository(test *testing.T) (*Repository, string) {
	test.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		test.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		test.Fatalf("open: %v", err)
	}
	test.Cleanup(pool.Close)

	migrations, err := database.Embedded()
	if err != nil {
		test.Fatalf("migrations: %v", err)
	}
	if _, err := database.Migrate(ctx, pool, migrations); err != nil {
		test.Fatalf("migrate: %v", err)
	}

	ownerID := "test-" + test.Name() + "-" + time.Now().Format("150405.000000000")
	_, err = pool.Exec(ctx, `
WITH created AS (
  INSERT INTO artifacts (owner_id, label, title, kind, format, tags, content_text)
  VALUES ($1, 'abertura-cnpj', 'Abertura do CNPJ', 'plan', 'md', '{empresa}', 'contador e junta comercial')
  RETURNING id
)
INSERT INTO artifact_versions (artifact_id, version, size_bytes, has_thumb)
SELECT created.id, 1, 2048, TRUE
FROM created`, ownerID)
	if err != nil {
		test.Fatalf("seed: %v", err)
	}
	test.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM artifacts WHERE owner_id = $1`, ownerID)
	})

	return NewRepository(pool), ownerID
}

// feliz
func TestListFindsByKindAndPortugueseSearch(test *testing.T) {
	repository, ownerID := seededRepository(test)

	found, err := repository.List(context.Background(), ownerID, Filter{Kind: "plan", Search: "junta comercial"})
	if err != nil {
		test.Fatalf("list: %v", err)
	}
	if len(found) != 1 || found[0].SizeBytes != 2048 || found[0].ThumbURL == nil {
		test.Fatalf("list = %+v, want the seeded artifact with its served version", found)
	}
}

// feliz
func TestOpenMarksTheVisitAndListsVersions(test *testing.T) {
	repository, ownerID := seededRepository(test)

	detail, err := repository.Open(context.Background(), ownerID, "abertura-cnpj")
	if err != nil {
		test.Fatalf("open: %v", err)
	}
	if detail.LastViewedAt == nil || len(detail.Versions) != 1 {
		test.Fatalf("open = %+v, want lastViewedAt set and one version", detail)
	}
}

// feliz
func TestUpdateChangesFlagsWithoutTouchingUpdatedAt(test *testing.T) {
	repository, ownerID := seededRepository(test)
	ctx := context.Background()
	before, err := repository.Open(ctx, ownerID, "abertura-cnpj")
	if err != nil {
		test.Fatalf("open: %v", err)
	}

	pinned, done := true, "done"
	updated, err := repository.Update(ctx, ownerID, "abertura-cnpj", Patch{IsPinned: &pinned, StatusSet: true, Status: &done})
	if err != nil {
		test.Fatalf("update: %v", err)
	}
	if !updated.IsPinned || updated.Status == nil || *updated.Status != "done" {
		test.Fatalf("update = %+v, want pinned and done", updated)
	}
	if !updated.UpdatedAt.Equal(before.UpdatedAt) {
		test.Fatal("pinning must not change updatedAt")
	}
}

// triste
func TestArtifactsOfAnotherOwnerAreNotFound(test *testing.T) {
	repository, ownerID := seededRepository(test)
	ctx := context.Background()

	if _, err := repository.Open(ctx, "someone-else", "abertura-cnpj"); !errors.Is(err, database.ErrNotFound) {
		test.Fatalf("open: err = %v, want ErrNotFound", err)
	}
	if _, err := repository.Update(ctx, "someone-else", "abertura-cnpj", Patch{}); !errors.Is(err, database.ErrNotFound) {
		test.Fatalf("update: err = %v, want ErrNotFound", err)
	}

	found, err := repository.List(ctx, ownerID, Filter{Search: "nada parecido aqui"})
	if err != nil || len(found) != 0 {
		test.Fatalf("search miss: list = %+v, err = %v", found, err)
	}
}
