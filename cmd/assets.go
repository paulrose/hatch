package cmd

import "embed"

var embeddedAssets embed.FS
var appIconData []byte

// SetAssets stores the embedded frontend filesystem for use by the app command.
func SetAssets(assets embed.FS) {
	embeddedAssets = assets
}

// SetAppIcon stores the embedded application icon for use by the app command.
func SetAppIcon(icon []byte) {
	appIconData = icon
}
