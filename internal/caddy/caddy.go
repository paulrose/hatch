// Package caddy translates Hatch configuration into Caddy's JSON
// configuration format and provides a client for the Caddy admin API.
package caddy

// DefaultAdminAddr is the default Caddy admin API listen address.
const DefaultAdminAddr = "localhost:2019"

// ServerConfig holds the settings needed to run the embedded Caddy server.
// It is decoupled from config.Settings — the caller maps between them.
type ServerConfig struct {
	AdminAddr string // e.g. "localhost:2019"
}
