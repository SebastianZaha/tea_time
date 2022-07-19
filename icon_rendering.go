package main

import (
	"bufio"
	"bytes"
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
	dpi      float64 = 200                             // screen resolution in Dots Per Inch
	fontfile         = "./ui/fonts/Oswald-Regular.ttf" // filename of the ttf font
	hinting          = "none"                          // "none | full"
	size     float64 = 14                              // font size in points
)

func loadFontFace() font.Face {
	// Read the font data.
	fontBytes, err := ioutil.ReadFile(fontfile)
	if err != nil {
		log.Println(err)
		return nil
	}
	f, err := truetype.Parse(fontBytes)
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
