package cmd

import (
	"fmt"
	"io/fs"

	"github.com/paulrose/hatch/internal/app"
	"github.com/spf13/cobra"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Launch the Hatch GUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := app.NewApp(version)

		frontendAssets, err := fs.Sub(embeddedAssets, "frontend/dist")
		if err != nil {
			return fmt.Errorf("loading frontend assets: %w", err)
		}

		return wails.Run(&options.App{
			Title:     "Hatch",
			Width:     1024,
			Height:    768,
			MinWidth:  800,
			MinHeight: 600,
			AssetServer: &assetserver.Options{
				Assets: frontendAssets,
			},
			OnStartup:  a.Startup,
			OnShutdown: a.Shutdown,
			Bind: []interface{}{
				a,
			},
		})
	},
}

func init() {
	rootCmd.AddCommand(appCmd)
}
