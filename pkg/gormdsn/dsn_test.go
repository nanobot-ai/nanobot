package gormdsn

import (
	"path/filepath"
	"testing"
)

func TestNewDBFromDSNSQLiteFormats(t *testing.T) {
	tests := map[string]string{
		"scheme":         "sqlite:file:",
		"absolute URL":   "sqlite://",
		"bare file path": "",
	}

	for name, prefix := range tests {
		t.Run(name, func(t *testing.T) {
			databasePath := filepath.Join(t.TempDir(), "nanobot.db")
			db, err := NewDBFromDSN(prefix + databasePath)
			if err != nil {
				t.Fatalf("NewDBFromDSN: %v", err)
			}
			if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)").Error; err != nil {
				t.Fatalf("create table: %v", err)
			}
		})
	}

	t.Run("memory", func(t *testing.T) {
		db, err := NewDBFromDSN("sqlite::memory:")
		if err != nil {
			t.Fatalf("NewDBFromDSN: %v", err)
		}
		if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)").Error; err != nil {
			t.Fatalf("create table: %v", err)
		}
	})
}
