//go:build !darwin

package tray

func setDockVisible(_ bool) {}

func setAppIcon(_ []byte) {}
