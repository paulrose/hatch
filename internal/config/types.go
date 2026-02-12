package config

// Config is the top-level Hatch configuration.
type Config struct {
	Version  int                `yaml:"version"`
	Settings Settings           `yaml:"settings"`
	Projects map[string]Project `yaml:"projects"`
}

// Settings holds global Hatch settings.
type Settings struct {
	TLD       string `yaml:"tld"`
	HTTPPort  int    `yaml:"http_port"`
	HTTPSPort int    `yaml:"https_port"`
	AutoStart bool   `yaml:"auto_start"`
	LogLevel  string `yaml:"log_level"`
}

// Project defines a single project's proxy configuration.
type Project struct {
	Domain   string             `yaml:"domain"`
	Path     string             `yaml:"path"`
	Enabled  bool               `yaml:"enabled"`
	Services map[string]Service `yaml:"services"`
}

// Service defines how a single service is proxied.
type Service struct {
	Proxy     string `yaml:"proxy"`
	Route     string `yaml:"route,omitempty"`
	Subdomain string `yaml:"subdomain,omitempty"`
	WebSocket bool   `yaml:"websocket,omitempty"`
}

// ProjectConfig is the schema for a per-project .hatch.yml file.
type ProjectConfig struct {
	Domain   string             `yaml:"domain"`
	Services map[string]Service `yaml:"services"`
}
