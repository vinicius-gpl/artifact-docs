package database

import (
	"strings"
	"testing"
	"testing/fstest"
)

// feliz
func TestListOrdersByNumberAcrossFolders(test *testing.T) {
	source := fstest.MapFS{
		"views/0003_create_summary_view.sql":    {Data: []byte("CREATE VIEW summary_view AS SELECT 1;")},
		"models/apps/0002_create_apps.sql":      {Data: []byte("CREATE TABLE apps ();")},
		"models/docs/0001_create_artifacts.sql": {Data: []byte("CREATE TABLE artifacts ();")},
	}

	migrations, err := List(source)
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	names := make([]string, 0, len(migrations))
	for _, migration := range migrations {
		names = append(names, migration.Name)
	}
	want := "0001_create_artifacts.sql 0002_create_apps.sql 0003_create_summary_view.sql"
	if strings.Join(names, " ") != want {
		test.Fatalf("order = %v, want %s (the folder must not matter)", names, want)
	}
	if migrations[0].SQL != "CREATE TABLE artifacts ();" {
		test.Fatalf("sql = %q", migrations[0].SQL)
	}
}

// feliz
func TestTheEmbeddedMigrationsAreValid(test *testing.T) {
	migrations, err := Embedded()
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if len(migrations) == 0 || migrations[0].Name != "0001_create_artifacts.sql" {
		test.Fatalf("migrations = %+v, want 0001_create_artifacts.sql first", migrations)
	}
	if !strings.Contains(migrations[0].SQL, "CREATE TABLE artifacts") {
		test.Fatal("0001 must create the artifacts table")
	}
}

// triste
func TestListRefusesAFileOutsideTheNamingRule(test *testing.T) {
	for _, name := range []string{"001_short.sql", "0001-dash.sql", "0001_Upper.sql", "models/0001_x.txt"} {
		source := fstest.MapFS{name: {Data: []byte("SELECT 1;")}}

		if _, err := List(source); err == nil {
			test.Errorf("%s: expected an error", name)
		}
	}
}

// triste
func TestListRefusesTwoMigrationsWithTheSameNumber(test *testing.T) {
	source := fstest.MapFS{
		"models/artifacts/0007_create_tags.sql": {Data: []byte("SELECT 1;")},
		"models/apps/0007_create_icons.sql":     {Data: []byte("SELECT 1;")},
	}

	_, err := List(source)
	if err == nil || !strings.Contains(err.Error(), "0007") {
		test.Fatalf("err = %v, want a duplicated-number error", err)
	}
}
