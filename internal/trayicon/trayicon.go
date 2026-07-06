// Package trayicon renders the tray/menu-bar icon as PNG bytes. The symbol is a
// simple horizontal ruler (placeholder for the FontAwesome ruler-horizontal
// asset, ADR-0005 / OBS-01) tinted by state color. Five states: gray (idle),
// blue (scanning), green (ok), yellow (warn), red (over).
package trayicon

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

type State int

const (
	Idle State = iota
	Scanning
	OK
	Warn
	Over
)

var colors = map[State]color.RGBA{
	Idle:     {0x9a, 0xa0, 0xa6, 0xff}, // gray
	Scanning: {0x1a, 0x73, 0xe8, 0xff}, // blue
	OK:       {0x34, 0xa8, 0x53, 0xff}, // green
	Warn:     {0xfb, 0xbc, 0x04, 0xff}, // amber
	Over:     {0xea, 0x43, 0x35, 0xff}, // red
}

// PNG returns the icon for a state as PNG bytes.
func PNG(s State) []byte {
	c, ok := colors[s]
	if !ok {
		c = colors[Idle]
	}
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Ruler body: a horizontal bar.
	const top, bottom, left, right = 11, 21, 3, 29
	for y := top; y <= bottom; y++ {
		for x := left; x <= right; x++ {
			img.Set(x, y, c)
		}
	}
	// Tick notches cut from the top edge, alternating depth — reads as a ruler.
	transparent := color.RGBA{}
	for i, x := 0, left+3; x < right; i, x = i+1, x+3 {
		depth := 4
		if i%2 == 1 {
			depth = 2
		}
		for y := top; y < top+depth; y++ {
			img.Set(x, y, transparent)
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
