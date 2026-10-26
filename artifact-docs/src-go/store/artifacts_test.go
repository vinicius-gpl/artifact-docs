package store

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/migrations"
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

// openTestStore conecta ao banco de teste, ou pula: o CI não tem banco, e o SQL de verdade
// só é conferido onde a TEST_DATABASE_URL existe (src-go/.env).
func openTestStore(test *testing.T) *Store {
	test.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		test.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	database, err := Open(ctx, databaseURL)
	if err != nil {
		test.Fatalf("open: %v", err)
	}
	test.Cleanup(database.Close)

	pending, err := migrations.Embedded()
	if err != nil {
		test.Fatalf("migrations: %v", err)
	}
	if _, err := migrations.Apply(ctx, database, pending); err != nil {
		test.Fatalf("apply: %v", err)
	}

	return database
}

// seedArtifact cria um artefato de um dono só deste teste, e o apaga no fim.
func seedArtifact(test *testing.T, database *Store, ownerID string) {
	test.Helper()

	ctx := context.Background()
	_, err := database.pool.Exec(ctx, `with created as (
	insert into artifacts (owner_id, label, title, kind, format, tags, content_text)
	values ($1, 'abertura-cnpj', 'Abertura do CNPJ', 'plan', 'md', '{empresa}', 'contador e junta comercial')
	returning id
)
insert into artifact_versions (artifact_id, version, size_bytes, has_thumb)
select id, 1, 2048, true from created`, ownerID)
	if err != nil {
		test.Fatalf("seed: %v", err)
	}

	test.Cleanup(func() {
		_, _ = database.pool.Exec(context.Background(), "delete from artifacts where owner_id = $1", ownerID)
	})
}

// seededStore é o banco de teste com um artefato de um dono que só este teste usa.
func seededStore(test *testing.T) (*Store, string) {
	test.Helper()

	database := openTestStore(test)
	ownerID := "test-" + test.Name() + "-" + time.Now().Format("150405.000000000")
	seedArtifact(test, database, ownerID)

	return database, ownerID
}

// feliz
func TestListArtifactsFindsByKindAndPortugueseSearch(test *testing.T) {
	database, ownerID := seededStore(test)

	found, err := database.ListArtifacts(context.Background(), ownerID, ArtifactFilter{Kind: "plan", Search: "junta comercial"})
	if err != nil {
		test.Fatalf("list: %v", err)
	}
	if len(found) != 1 || found[0].SizeBytes != 2048 || found[0].ThumbURL == nil {
		test.Fatalf("list = %+v, want the seeded artifact with its served version", found)
	}
}

// feliz
func TestOpenArtifactMarksTheVisitAndListsVersions(test *testing.T) {
	database, ownerID := seededStore(test)

	detail, err := database.OpenArtifact(context.Background(), ownerID, "abertura-cnpj")
	if err != nil {
		test.Fatalf("open: %v", err)
	}
	if detail.LastViewedAt == nil || len(detail.Versions) != 1 {
		test.Fatalf("open = %+v, want lastViewedAt set and one version", detail)
	}
}

// feliz
func TestPatchArtifactChangesFlagsWithoutTouchingUpdatedAt(test *testing.T) {
	database, ownerID := seededStore(test)
	ctx := context.Background()
	before, err := database.OpenArtifact(ctx, ownerID, "abertura-cnpj")
	if err != nil {
		test.Fatalf("open: %v", err)
	}

	pinned, done := true, "done"
	patched, err := database.PatchArtifact(ctx, ownerID, "abertura-cnpj", ArtifactPatch{IsPinned: &pinned, StatusSet: true, Status: &done})
	if err != nil {
		test.Fatalf("patch: %v", err)
	}
	if !patched.IsPinned || patched.Status == nil || *patched.Status != "done" {
		test.Fatalf("patch = %+v, want pinned and done", patched)
	}
	if !patched.UpdatedAt.Equal(before.UpdatedAt) {
		test.Fatal("pinning must not change updatedAt")
	}
}

// triste
func TestArtifactsOfAnotherOwnerAreNotFound(test *testing.T) {
	database, ownerID := seededStore(test)
	ctx := context.Background()

	if _, err := database.OpenArtifact(ctx, "someone-else", "abertura-cnpj"); !errors.Is(err, ErrNotFound) {
		test.Fatalf("open: err = %v, want ErrNotFound", err)
	}
	if _, err := database.PatchArtifact(ctx, "someone-else", "abertura-cnpj", ArtifactPatch{}); !errors.Is(err, ErrNotFound) {
		test.Fatalf("patch: err = %v, want ErrNotFound", err)
	}

	found, err := database.ListArtifacts(ctx, ownerID, ArtifactFilter{Search: "nada parecido aqui"})
	if err != nil || len(found) != 0 {
		test.Fatalf("search miss: list = %+v, err = %v", found, err)
	}
}
