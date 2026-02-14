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

// Pre-rendered template icon variants as PNG bytes.
// All icons are alpha-only (black on transparent) so macOS tints them
// automatically for dark/light mode when used with SetTemplateIcon.
var (
	IconHealthy  []byte // Clean icon, no badge
	IconDegraded []byte // Icon + filled dot at bottom-right
	IconError    []byte // Icon + ring at bottom-right
	IconStopped  []byte // Icon at 50% alpha
)

func init() {
	base, err := png.Decode(bytes.NewReader(iconPNG))
	if err != nil {
		panic("tray: failed to decode embedded icon.png: " + err.Error())
	}

	IconHealthy = encodePNG(base)
	IconDegraded = renderDotBadge(base)
	IconError = renderRingBadge(base)
	IconStopped = renderDimmed(base)
}

func encodePNG(img image.Image) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return buf.Bytes()
}

// renderDotBadge draws a filled 10px circle at the bottom-right.
func renderDotBadge(base image.Image) []byte {
	bounds := base.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, base, bounds.Min, draw.Src)

	radius := 5
	cx := bounds.Max.X - radius - 2
	cy := bounds.Max.Y - radius - 2
	black := color.RGBA{A: 255}

	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, black)
			}
		}
	}

	return encodePNG(img)
}

// renderRingBadge draws a 10px ring (2px stroke, hollow) at the bottom-right.
func renderRingBadge(base image.Image) []byte {
	bounds := base.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, base, bounds.Min, draw.Src)

	radius := 5
	stroke := 2
	cx := bounds.Max.X - radius - 2
	cy := bounds.Max.Y - radius - 2
	black := color.RGBA{A: 255}
	inner := radius - stroke

	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx := x - cx
			dy := y - cy
			distSq := dx*dx + dy*dy
			if distSq <= radius*radius && distSq > inner*inner {
				img.Set(x, y, black)
			}
		}
	}

	return encodePNG(img)
}

// renderDimmed returns the base icon with all alpha values halved.
func renderDimmed(base image.Image) []byte {
	bounds := base.Bounds()
	img := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := base.At(x, y).RGBA()
			img.Set(x, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8((a >> 8) / 2),
			})
		}
	}

	return encodePNG(img)
}
