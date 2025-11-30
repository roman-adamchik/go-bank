package util

import (
	"os"
	"path/filepath"
	"testing"
)

// Test successful load from app.env
func TestLoadConfig_Success(t *testing.T) {
	dir := t.TempDir()
	content := "DB_SOURCE=postgresql://user:pass@localhost:5432/db\nSERVER_ADDRESS=0.0.0.0:8080\n"
	if err := os.WriteFile(filepath.Join(dir, "app.env"), []byte(content), 0o600); err != nil {
		t.Fatalf("write app.env failed: %v", err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.DBSource != "postgresql://user:pass@localhost:5432/db" {
		t.Fatalf("unexpected DBSource: %s", cfg.DBSource)
	}
	if cfg.ServerAddress != "0.0.0.0:8080" {
		t.Fatalf("unexpected ServerAddress: %s", cfg.ServerAddress)
	}
}

// Test error when config file missing
func TestLoadConfig_FileMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatalf("expected error for missing config file, got nil")
	}
}
