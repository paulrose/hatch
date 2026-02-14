package tray

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

//go:embed icon.png
var iconPNG []byte

// Badge colors matching the original ObjC values.
var (
	colorGreen  = color.RGBA{R: 77, G: 199, B: 102, A: 255}
	colorYellow = color.RGBA{R: 242, G: 196, B: 15, A: 255}
	colorRed    = color.RGBA{R: 232, G: 77, B: 61, A: 255}
	colorGray   = color.RGBA{R: 153, G: 153, B: 153, A: 255}
)

// Pre-rendered badge icon variants as PNG bytes.
var (
	BadgeGreen  []byte
	BadgeYellow []byte
	BadgeRed    []byte
	BadgeGray   []byte
)

func init() {
	base, err := png.Decode(bytes.NewReader(iconPNG))
	if err != nil {
		panic("tray: failed to decode embedded icon.png: " + err.Error())
	}

	BadgeGreen = renderBadge(base, colorGreen)
	BadgeYellow = renderBadge(base, colorYellow)
	BadgeRed = renderBadge(base, colorRed)
	BadgeGray = renderBadge(base, colorGray)
}

// renderBadge draws a filled circle at the bottom-right of the base icon.
func renderBadge(base image.Image, dotColor color.Color) []byte {
	bounds := base.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, base, bounds.Min, draw.Src)

	// Draw a 12px filled circle at bottom-right with 2px inset.
	dotRadius := 6
	cx := bounds.Max.X - dotRadius - 2
	cy := bounds.Max.Y - dotRadius - 2

	for y := cy - dotRadius; y <= cy+dotRadius; y++ {
		for x := cx - dotRadius; x <= cx+dotRadius; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= dotRadius*dotRadius {
				img.Set(x, y, dotColor)
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return buf.Bytes()
}
