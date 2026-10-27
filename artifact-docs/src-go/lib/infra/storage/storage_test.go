package storage

import (
	"errors"
	"testing"

	"github.com/minio/minio-go/v7"
)

// feliz
func TestOpenAcceptsTheR2Endpoint(test *testing.T) {
	bucket, err := Open(Settings{
		Endpoint:        "https://account.r2.cloudflarestorage.com",
		AccessKeyID:     "key",
		SecretAccessKey: "secret",
		Bucket:          "artifact-docs",
	})
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if bucket.name != "artifact-docs" {
		test.Fatalf("bucket = %q", bucket.name)
	}
}

// triste
func TestOpenRefusesAnEndpointThatIsNotAURL(test *testing.T) {
	for _, endpoint := range []string{"", "account.r2", "::"} {
		if _, err := Open(Settings{Endpoint: endpoint}); err == nil {
			test.Errorf("endpoint %q: expected an error", endpoint)
		}
	}
}

// feliz
func TestAMissingKeyBecomesErrNotFound(test *testing.T) {
	err := notFoundOr(minio.ErrorResponse{Code: minio.NoSuchKey})

	if !errors.Is(err, ErrNotFound) {
		test.Fatalf("err = %v, want ErrNotFound", err)
	}
}

// triste
func TestOtherStorageErrorsPassThrough(test *testing.T) {
	original := minio.ErrorResponse{Code: "AccessDenied"}

	if err := notFoundOr(original); errors.Is(err, ErrNotFound) {
		test.Fatal("access denied must not look like a missing file")
	}
}
