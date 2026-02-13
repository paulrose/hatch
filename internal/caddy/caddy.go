// Package caddy translates Hatch configuration into Caddy's JSON
// configuration format and provides a client for the Caddy admin API.
package caddy

import (
	"os"
	"path/filepath"

	caddyv2 "github.com/caddyserver/caddy/v2"
)

// DefaultAdminAddr is the default Caddy admin API listen address.
const DefaultAdminAddr = "localhost:2019"

// ServerConfig holds the settings needed to run the embedded Caddy server.
// It is decoupled from config.Settings — the caller maps between them.
type ServerConfig struct {
	AdminAddr string // e.g. "localhost:2019"
}

// DataDir returns Caddy's application data directory.
func DataDir() string {
	return caddyv2.AppDataDir()
}

// ClearPKICache removes Caddy's cached PKI authority and issued certificates
// for the "hatch" CA. This forces Caddy to use the intermediate CA we provide
// in the PKI config and re-issue leaf certificates signed by it.
func ClearPKICache() error {
	dataDir := caddyv2.AppDataDir()
	dirs := []string{
		filepath.Join(dataDir, "pki", "authorities", "hatch"),
		filepath.Join(dataDir, "certificates", "hatch"),
	}
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	return nil
}
