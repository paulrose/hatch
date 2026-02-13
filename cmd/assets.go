package cmd

import "embed"

var embeddedAssets embed.FS

// SetAssets stores the embedded frontend filesystem for use by the app command.
func SetAssets(assets embed.FS) {
	embeddedAssets = assets
}
