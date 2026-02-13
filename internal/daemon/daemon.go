// Package daemon manages the Hatch background process lifecycle,
// including start, stop, and status operations.
package daemon

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/paulrose/hatch/internal/caddy"
	"github.com/paulrose/hatch/internal/certs"
	"github.com/paulrose/hatch/internal/config"
	"github.com/paulrose/hatch/internal/dns"
	"github.com/paulrose/hatch/internal/health"
)

// Daemon orchestrates all Hatch subsystems as a long-running background process.
type Daemon struct {
	caddy   *caddy.Server
	dns     *dns.Server
	health  *health.Checker
	watcher *config.Watcher
	pidFile *os.File
	caPaths certs.CAPaths
	cfg     config.Config
	mu      sync.Mutex
	running bool
}

// New creates a new Daemon instance.
func New() *Daemon {
	return &Daemon{}
}

// Run starts all subsystems and blocks until ctx is cancelled.
// On context cancellation it performs a graceful shutdown.
func (d *Daemon) Run(ctx context.Context) error {
	// Write PID file (acquires flock).
	pidFile, err := WritePID()
	if err != nil {
		return fmt.Errorf("write pid: %w", err)
	}
	d.pidFile = pidFile
	log.Info().Int("pid", os.Getpid()).Msg("pid file written")

	// Load config.
	cfg, err := config.Load()
	if err != nil {
		RemovePID(d.pidFile)
		return fmt.Errorf("load config: %w", err)
	}
	d.cfg = cfg

	// Resolve CA paths and verify the files exist.
	d.caPaths = certs.NewCAPaths(config.CertsDir())
	if !certs.CAExists(d.caPaths) {
		RemovePID(d.pidFile)
		return fmt.Errorf("CA files not found at %s — run 'hatch up' to generate", config.CertsDir())
	}

	// Start DNS server.
	dnsSrv, err := dns.NewServer(dns.ServerConfig{
		TLD:      cfg.Settings.TLD,
		ListenIP: dns.DefaultListenIP,
		Port:     dns.DefaultPort,
	})
	if err != nil {
		RemovePID(d.pidFile)
		return fmt.Errorf("create dns server: %w", err)
	}
	if err := dnsSrv.Start(); err != nil {
		RemovePID(d.pidFile)
		return fmt.Errorf("start dns: %w", err)
	}
	d.dns = dnsSrv
	log.Info().Str("tld", cfg.Settings.TLD).Int("port", dns.DefaultPort).Msg("dns server started")

	// Start Caddy server.
	caddySrv := caddy.NewServer(caddy.ServerConfig{
		AdminAddr: caddy.DefaultAdminAddr,
	})
	if err := caddySrv.Start(ctx); err != nil {
		d.shutdownPartial()
		return fmt.Errorf("start caddy: %w", err)
	}
	d.caddy = caddySrv
	log.Info().Msg("caddy server started")

	// Load translated config into Caddy.
	caddyCfg := caddy.Translate(cfg, d.caPaths.Cert, d.caPaths.Key)
	if err := caddySrv.LoadConfig(ctx, caddyCfg); err != nil {
		d.shutdownPartial()
		return fmt.Errorf("load caddy config: %w", err)
	}
	log.Info().Msg("caddy config loaded")

	// Start health checker.
	checker := health.NewChecker(health.CheckerConfig{})
	if err := checker.Start(cfg); err != nil {
		d.shutdownPartial()
		return fmt.Errorf("start health checker: %w", err)
	}
	d.health = checker
	log.Info().Msg("health checker started")

	// Start config watcher.
	watcher, err := config.NewWatcher(d.onConfigReload)
	if err != nil {
		d.shutdownPartial()
		return fmt.Errorf("start config watcher: %w", err)
	}
	d.watcher = watcher
	log.Info().Msg("config watcher started")

	d.mu.Lock()
	d.running = true
	d.mu.Unlock()

	log.Info().Msg("daemon running")

	// Block until context is cancelled.
	<-ctx.Done()
	log.Info().Msg("shutdown signal received")

	return d.Shutdown()
}

// Shutdown stops all subsystems in reverse start order and removes the PID file.
func (d *Daemon) Shutdown() error {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return nil
	}
	d.running = false
	d.mu.Unlock()

	var errs []error

	// Watcher first — prevents reload during teardown.
	if d.watcher != nil {
		if err := d.watcher.Close(); err != nil {
			errs = append(errs, fmt.Errorf("stop watcher: %w", err))
		}
		log.Info().Msg("config watcher stopped")
	}

	// Health checker.
	if d.health != nil {
		if err := d.health.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("stop health checker: %w", err))
		}
		log.Info().Msg("health checker stopped")
	}

	// Caddy.
	if d.caddy != nil {
		if err := d.caddy.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("stop caddy: %w", err))
		}
		log.Info().Msg("caddy server stopped")
	}

	// DNS.
	if d.dns != nil {
		if err := d.dns.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("stop dns: %w", err))
		}
		log.Info().Msg("dns server stopped")
	}

	// Remove PID file.
	if d.pidFile != nil {
		if err := RemovePID(d.pidFile); err != nil {
			errs = append(errs, fmt.Errorf("remove pid: %w", err))
		}
		log.Info().Msg("pid file removed")
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	log.Info().Msg("daemon stopped")
	return nil
}

// onConfigReload is called by the config watcher when the config file changes.
// It re-translates the Caddy config and updates the health checker.
func (d *Daemon) onConfigReload(cfg config.Config) {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return
	}
	d.cfg = cfg
	d.mu.Unlock()

	caddyCfg := caddy.Translate(cfg, d.caPaths.Cert, d.caPaths.Key)
	if err := d.caddy.LoadConfig(context.Background(), caddyCfg); err != nil {
		log.Error().Err(err).Msg("failed to reload caddy config")
		return
	}

	d.health.UpdateConfig(cfg)
	log.Info().Msg("config reloaded successfully")
}

// shutdownPartial stops any subsystems that were started during a failed Run.
func (d *Daemon) shutdownPartial() {
	if d.health != nil {
		d.health.Stop()
	}
	if d.caddy != nil {
		d.caddy.Stop()
	}
	if d.dns != nil {
		d.dns.Stop()
	}
	if d.pidFile != nil {
		RemovePID(d.pidFile)
	}
}
