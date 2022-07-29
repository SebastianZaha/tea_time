package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// https://stackoverflow.com/questions/38299930/how-to-add-a-simple-text-label-to-an-image-in-go
// https://github.com/golang/freetype/blob/master/example/drawer/main.go
var (
	dpi     float64 = 200    // screen resolution in Dots Per Inch
	hinting         = "none" // "none | full"
	size    float64 = 14     // font size in points

)

//go:embed ui/fonts/Oswald-Regular.ttf
var fontfile []byte

var (
	colorHour = parseHexColor("#9afcba")
	colorMin  = parseHexColor("#faefa3")
	colorSec  = parseHexColor("#ff978a")
)

func loadFontFace() font.Face {
	f, err := truetype.Parse(fontfile)
	if err != nil {
		log.Println(err)
		return nil
	}

	h := font.HintingNone
	switch hinting {
	case "full":
		h = font.HintingFull
	}

	return truetype.NewFace(f, &truetype.Options{
		Size:    size,
		DPI:     dpi,
		Hinting: h,
	})
}

var fontFace = loadFontFace()

func drawText(text string, color color.Color) []byte {
	const imgW, imgH = 50, 50
	rgba := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	// background? draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color),
		Face: fontFace,
	}

	x := (fixed.I(imgW) - d.MeasureString(text)) / 2
	// 4 magic here accounts for our characters always being smaller than the font height
	fontH := int(math.Ceil(size*dpi/72)) - 4
	y := (fixed.I(imgH) + fixed.I(fontH)) / 2
	d.Dot = fixed.Point26_6{x, y}
	d.DrawString(text)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := png.Encode(w, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = w.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	bs, err := ioutil.ReadAll(&b)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	return bs
}

// https://stackoverflow.com/a/54200713/1306453
func parseHexColor(s string) (c color.RGBA) {
	c.A = 0xff

	if s[0] != '#' {
		log.Fatalf("invalid color format %v, no '#' at start", s)
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		log.Fatalf("invalid color format %v, includes non hex char %v", s, b)
		return 0
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	default:
		log.Fatalf("invalid color format %v, should have 7 or 4 characters", s)
	}
	return
}
