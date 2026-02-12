package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HATCH_HOME", dir)
	return dir
}

func TestLoadSaveRoundTrip(t *testing.T) {
	setupTestHome(t)
	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	cfg.Projects["myapp"] = Project{
		Domain: "myapp.test", Path: "/tmp/myapp", Enabled: true,
		Services: map[string]Service{"web": {Proxy: "http://localhost:3000"}},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Version != cfg.Version {
		t.Errorf("version: got %d, want %d", loaded.Version, cfg.Version)
	}
	if loaded.Settings.TLD != cfg.Settings.TLD {
		t.Errorf("tld: got %q, want %q", loaded.Settings.TLD, cfg.Settings.TLD)
	}
	p, ok := loaded.Projects["myapp"]
	if !ok {
		t.Fatal("project myapp not found after round-trip")
	}
	if p.Domain != "myapp.test" {
		t.Errorf("domain: got %q, want %q", p.Domain, "myapp.test")
	}
	if p.Services["web"].Proxy != "http://localhost:3000" {
		t.Errorf("proxy: got %q, want %q", p.Services["web"].Proxy, "http://localhost:3000")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	setupTestHome(t)
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	home := setupTestHome(t)
	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, configFileName), []byte(":::bad"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoad_InvalidConfig(t *testing.T) {
	home := setupTestHome(t)
	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}
	// Version 0 is invalid
	if err := os.WriteFile(filepath.Join(home, configFileName), []byte("version: 0\nsettings:\n  tld: test\n  http_port: 80\n  https_port: 443\n  log_level: info\nprojects: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	home := setupTestHome(t)
	if err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Ensure no .tmp file remains
	tmp := filepath.Join(home, configFileName+".tmp")
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Error("temp file should not exist after successful save")
	}

	// Ensure the actual file exists
	if _, err := os.Stat(filepath.Join(home, configFileName)); err != nil {
		t.Errorf("config file should exist: %v", err)
	}
}

func TestLoadProjectConfig_Valid(t *testing.T) {
	path := filepath.Join("testdata", "project.yml")
	pc, err := LoadProjectConfig(path)
	if err != nil {
		t.Fatalf("LoadProjectConfig: %v", err)
	}
	if pc.Domain != "myapp.test" {
		t.Errorf("domain: got %q, want %q", pc.Domain, "myapp.test")
	}
	if _, ok := pc.Services["web"]; !ok {
		t.Error("expected web service")
	}
}

func TestLoadProjectConfig_MissingDomain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hatch.yml")
	if err := os.WriteFile(path, []byte("services:\n  web:\n    proxy: http://localhost:3000\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadProjectConfig(path)
	if err == nil {
		t.Fatal("expected error for missing domain")
	}
}

func TestLoadProjectConfig_NoServices(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hatch.yml")
	if err := os.WriteFile(path, []byte("domain: myapp.test\nservices: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadProjectConfig(path)
	if err == nil {
		t.Fatal("expected error for empty services")
	}
}

func TestLoadRaw(t *testing.T) {
	setupTestHome(t)
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadRaw()
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if cfg.Projects == nil {
		t.Error("Projects should be initialized to non-nil map")
	}
}
