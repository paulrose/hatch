package app

import "context"

// App is the Wails application struct, bound to the frontend via wails.Run.
type App struct {
	ctx     context.Context
	version string
}

func NewApp(version string) *App {
	return &App{version: version}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown(_ context.Context) {}

// GetVersion returns the application version string.
// Exposed to the frontend via Wails bindings.
func (a *App) GetVersion() string {
	return a.version
}
