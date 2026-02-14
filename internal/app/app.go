package app

// App is the Wails application service, exposed to the frontend via bindings.
type App struct {
	version string
}

func NewApp(version string) *App {
	return &App{version: version}
}

// GetVersion returns the application version string.
func (a *App) GetVersion() string {
	return a.version
}
