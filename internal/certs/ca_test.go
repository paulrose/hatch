package certs

import (
	"crypto/elliptic"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateCA(t *testing.T) {
	dir := t.TempDir()
	paths := NewCAPaths(dir)

	if err := GenerateCA(paths); err != nil {
		t.Fatalf("GenerateCA: %v", err)
	}

	// Both files should exist and be non-empty.
	for _, p := range []string{paths.Cert, paths.Key} {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("expected file %s to exist: %v", p, err)
		}
		if info.Size() == 0 {
			t.Errorf("expected file %s to be non-empty", p)
		}
	}
}

func TestGenerateCA_ValidCert(t *testing.T) {
	dir := t.TempDir()
	paths := NewCAPaths(dir)

	if err := GenerateCA(paths); err != nil {
		t.Fatalf("GenerateCA: %v", err)
	}

	cert, key, err := LoadCA(paths)
	if err != nil {
		t.Fatalf("LoadCA: %v", err)
	}

	if !cert.IsCA {
		t.Error("expected IsCA to be true")
	}
	if cert.Subject.CommonName != CACommonName {
		t.Errorf("expected CN %q, got %q", CACommonName, cert.Subject.CommonName)
	}
	if len(cert.Subject.Organization) == 0 || cert.Subject.Organization[0] != CAOrg {
		t.Errorf("expected O %q, got %v", CAOrg, cert.Subject.Organization)
	}
	if cert.KeyUsage&x509.KeyUsageCertSign == 0 {
		t.Error("expected KeyUsageCertSign")
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign == 0 {
		t.Error("expected KeyUsageCRLSign")
	}
	if cert.MaxPathLen != 1 {
		t.Errorf("expected MaxPathLen 1, got %d", cert.MaxPathLen)
	}

	// ECDSA P-256.
	if key.Curve != elliptic.P256() {
		t.Error("expected P-256 curve")
	}

	// Validity window.
	now := time.Now()
	if cert.NotBefore.After(now) {
		t.Error("cert NotBefore is in the future")
	}
	expectedExpiry := now.AddDate(CAValidYears, 0, 0)
	if cert.NotAfter.Before(expectedExpiry.Add(-time.Minute)) {
		t.Errorf("cert NotAfter %v is too early (expected around %v)", cert.NotAfter, expectedExpiry)
	}
}

func TestLoadCA(t *testing.T) {
	dir := t.TempDir()
	paths := NewCAPaths(dir)

	if err := GenerateCA(paths); err != nil {
		t.Fatalf("GenerateCA: %v", err)
	}

	cert, key, err := LoadCA(paths)
	if err != nil {
		t.Fatalf("LoadCA: %v", err)
	}

	// Verify the public keys match.
	if !key.PublicKey.Equal(cert.PublicKey) {
		t.Error("loaded key does not match certificate public key")
	}
}

func TestCAExists(t *testing.T) {
	dir := t.TempDir()
	paths := NewCAPaths(dir)

	if CAExists(paths) {
		t.Error("CAExists should be false before generation")
	}

	if err := GenerateCA(paths); err != nil {
		t.Fatalf("GenerateCA: %v", err)
	}

	if !CAExists(paths) {
		t.Error("CAExists should be true after generation")
	}

	// Remove cert file — should return false.
	os.Remove(paths.Cert)
	if CAExists(paths) {
		t.Error("CAExists should be false when cert is missing")
	}
}

func TestGenerateCA_MkdirCreatesParent(t *testing.T) {
	base := t.TempDir()
	nested := filepath.Join(base, "deep", "nested", "certs")
	paths := NewCAPaths(nested)

	if err := GenerateCA(paths); err != nil {
		t.Fatalf("GenerateCA with nested dir: %v", err)
	}

	if !CAExists(paths) {
		t.Error("expected CA files to exist in nested directory")
	}
}
