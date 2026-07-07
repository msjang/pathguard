// Command appicon renders the Pathguard app icon (a ruler on a rounded square)
// as a PNG. Used by scripts/make-icns.sh to build AppIcon.icns.
//
//	go run ./tools/appicon [size] [out.png]   (default 1024, appicon.png)
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"
)

func clamp(v float64) uint8 {
	switch {
	case v < 0:
		return 0
	case v > 255:
		return 255
	default:
		return uint8(v + 0.5)
	}
}

func blend(img *image.RGBA, x, y int, c color.RGBA, a float64) {
	o := img.RGBAAt(x, y)
	img.SetRGBA(x, y, color.RGBA{
		clamp(float64(c.R)*a + float64(o.R)*(1-a)),
		clamp(float64(c.G)*a + float64(o.G)*(1-a)),
		clamp(float64(c.B)*a + float64(o.B)*(1-a)),
		clamp(float64(c.A)*a + float64(o.A)*(1-a)),
	})
}

func coverage(px, py, x0, y0, x1, y1, r float64) float64 {
	corner := (px < x0+r || px > x1-r) && (py < y0+r || py > y1-r)
	var d float64
	if corner {
		dx := math.Max(x0+r-px, px-(x1-r))
		dy := math.Max(y0+r-py, py-(y1-r))
		d = math.Hypot(dx, dy) - r
	} else {
		d = math.Max(math.Max(x0-px, px-x1), math.Max(y0-py, py-y1))
	}
	switch {
	case d <= -0.5:
		return 1
	case d >= 0.5:
		return 0
	default:
		return 0.5 - d
	}
}

func fillRoundRect(img *image.RGBA, x0, y0, x1, y1, r float64, c color.RGBA) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if cov := coverage(float64(x)+0.5, float64(y)+0.5, x0, y0, x1, y1, r); cov > 0 {
				blend(img, x, y, c, cov)
			}
		}
	}
}

func rect(img *image.RGBA, x0, y0, x1, y1 float64, c color.RGBA) {
	for y := int(y0); y < int(y1); y++ {
		for x := int(x0); x < int(x1); x++ {
			blend(img, x, y, c, 1)
		}
	}
}

// Filled ruler body with graduation notches cut back to the background color.
func drawRuler(img *image.RGBA, S float64, body, bg color.RGBA) {
	bx0, bx1 := 0.13*S, 0.87*S
	by0, by1 := 0.40*S, 0.60*S
	rect(img, bx0, by0, bx1, by1, body)
	inner := by1 - by0
	long, short := 0.62*inner, 0.34*inner
	tw := 0.022 * S
	const n = 9
	span := bx1 - bx0
	for i := 1; i < n; i++ {
		tx := bx0 + span*float64(i)/float64(n)
		d := short
		if i%2 == 1 {
			d = long
		}
		rect(img, tx, by0, tx+tw, by0+d, bg)
	}
}

func downscale2(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx()/2, b.Dy()/2
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, bl, a float64
			for dy := 0; dy < 2; dy++ {
				for dx := 0; dx < 2; dx++ {
					p := src.RGBAAt(x*2+dx, y*2+dy)
					r, g, bl, a = r+float64(p.R), g+float64(p.G), bl+float64(p.B), a+float64(p.A)
				}
			}
			dst.SetRGBA(x, y, color.RGBA{clamp(r / 4), clamp(g / 4), clamp(bl / 4), clamp(a / 4)})
		}
	}
	return dst
}

func render(S int) *image.RGBA {
	bg := color.RGBA{0x1a, 0x73, 0xe8, 0xff}   // blue
	body := color.RGBA{0xff, 0xff, 0xff, 0xff} // white ruler
	ss := S * 2                                // supersample for anti-aliasing
	img := image.NewRGBA(image.Rect(0, 0, ss, ss))
	m := 0.075 * float64(ss)
	r := 0.225 * (float64(ss) - 2*m)
	fillRoundRect(img, m, m, float64(ss)-m, float64(ss)-m, r, bg)
	drawRuler(img, float64(ss), body, bg)
	return downscale2(img)
}

func main() {
	size := 1024
	out := "appicon.png"
	if len(os.Args) > 1 {
		if n, err := strconv.Atoi(os.Args[1]); err == nil {
			size = n
		}
	}
	if len(os.Args) > 2 {
		out = os.Args[2]
	}
	f, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, render(size)); err != nil {
		panic(err)
	}
}
