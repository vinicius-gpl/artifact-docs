package migrations

import (
	"strings"
	"testing"
	"testing/fstest"
)

// feliz
func TestListReturnsTheFilesInNumericOrder(test *testing.T) {
	source := fstest.MapFS{
		"0002_apps.sql":      {Data: []byte("create table apps ();")},
		"0001_artifacts.sql": {Data: []byte("create table artifacts ();")},
	}

	migrations, err := List(source)
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if len(migrations) != 2 || migrations[0].Name != "0001_artifacts.sql" || migrations[1].Name != "0002_apps.sql" {
		test.Fatalf("migrations = %+v, want 0001 then 0002", migrations)
	}
	if migrations[0].SQL != "create table artifacts ();" {
		test.Fatalf("sql = %q", migrations[0].SQL)
	}
}

// feliz
func TestTheEmbeddedMigrationsAreValid(test *testing.T) {
	migrations, err := Embedded()
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if len(migrations) == 0 || migrations[0].Name != "0001_artifacts.sql" {
		test.Fatalf("migrations = %+v, want 0001_artifacts.sql first", migrations)
	}
	if !strings.Contains(migrations[0].SQL, "create table artifacts") {
		test.Fatal("0001 must create the artifacts table")
	}
}

// triste
func TestListRefusesAFileOutsideTheNamingRule(test *testing.T) {
	for _, name := range []string{"001_short.sql", "0001-dash.sql", "0001_Upper.sql"} {
		source := fstest.MapFS{name: {Data: []byte("select 1;")}}

		if _, err := List(source); err == nil {
			test.Errorf("%s: expected an error", name)
		}
	}
}
