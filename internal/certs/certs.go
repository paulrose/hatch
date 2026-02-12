// Package certs handles root CA generation and macOS Keychain trust for
// local HTTPS development certificates.
package certs

import "path/filepath"

const (
	RootCACertFile = "rootCA.pem"
	RootCAKeyFile  = "rootCA-key.pem"
	CACommonName   = "Hatch Local CA"
	CAOrg          = "Hatch"
	CAValidYears   = 10
)

// CAPaths holds the file paths for the root CA certificate and key.
type CAPaths struct {
	Cert string // e.g. "~/.hatch/certs/rootCA.pem"
	Key  string // e.g. "~/.hatch/certs/rootCA-key.pem"
}

// NewCAPaths returns CAPaths rooted in certsDir.
func NewCAPaths(certsDir string) CAPaths {
	return CAPaths{
		Cert: filepath.Join(certsDir, RootCACertFile),
		Key:  filepath.Join(certsDir, RootCAKeyFile),
	}
}
