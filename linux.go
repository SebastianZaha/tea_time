//go:build linux

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"
)

func iconInactive() []byte {
	return drawIcon(13, false, 3, 15)
}

// font rendering

var glyphs = [][]byte{
	[]byte{0b0110, 0b1001, 0b1001, 0b1011, 0b1101, 0b1001, 0b1001, 0b0110}, // 0
	[]byte{0b0010, 0b0110, 0b0010, 0b0010, 0b0010, 0b0010, 0b0010, 0b0111}, // 1
	[]byte{0b0110, 0b1001, 0b0001, 0b0001, 0b0010, 0b0100, 0b1000, 0b1111}, // 2
	[]byte{0b1111, 0b0001, 0b0010, 0b0101, 0b0001, 0b0001, 0b1001, 0b0110}, // 3
	[]byte{0b1000, 0b1000, 0b1010, 0b1010, 0b1111, 0b0010, 0b0010, 0b0010},
	[]byte{0b1111, 0b1000, 0b1000, 0b1110, 0b0001, 0b0001, 0b1001, 0b0110}, // 5
	[]byte{0b0110, 0b1001, 0b1000, 0b1000, 0b1110, 0b1001, 0b1001, 0b0110},
	[]byte{0b1111, 0b0001, 0b0001, 0b0010, 0b0010, 0b0100, 0b0100, 0b0010}, // 7
	[]byte{0b0110, 0b1001, 0b1001, 0b0110, 0b0110, 0b1001, 0b1001, 0b0110},
	[]byte{0b0110, 0b1001, 0b1001, 0b0111, 0b0001, 0b0001, 0b1001, 0b0110}, // 9
	[]byte{0b0000, 0b0000, 0b0000, 0b1001, 0b1111, 0b1001, 0b1001, 0b1001}, // m 10
	[]byte{0b0000, 0b1000, 0b1000, 0b1110, 0b1001, 0b1001, 0b1001, 0b1001}, // h 11
	[]byte{0b0000, 0b0000, 0b0000, 0b0111, 0b1000, 0b0110, 0b0001, 0b1110}, // s 12
	[]byte{0b0000, 0b0100, 0b0100, 0b1111, 0b0100, 0b0100, 0b0100, 0b0011}, // t 13
	[]byte{0b0000, 0b0000, 0b0000, 0b0111, 0b1001, 0b1001, 0b1011, 0b0101}, // a 14
	[]byte{0b0000, 0b0000, 0b0000, 0b0000, 0b0000, 0b0000, 0b0000, 0b0000}, // space 15
}

func txtForDuration(d time.Duration) string {
	return ""
}

func iconForDuration(d time.Duration) []byte {
	if d >= 100*time.Hour {
		return drawIcon(11, false, 11, 11)
	} else if d >= 10*time.Hour {
		return drawIcon(byte(d/time.Hour/10), false, byte((d/time.Hour)%10), 11)
	} else if d > time.Hour {
		i, frac := math.Modf(d.Hours())
		return drawIcon(byte(i), true, byte(frac*10), 11)
	} else if d >= 10*time.Minute {
		return drawIcon(byte(d/time.Minute/10), false, byte((d/time.Minute)%10), 10)
	} else if d > time.Minute {
		i, frac := math.Modf(d.Minutes())
		return drawIcon(byte(i), true, byte(frac*10), 10)
	} else {
		return drawIcon(byte(d/time.Second/10), false, byte(d/time.Second%10), 12)
	}
}

func drawIcon(d1 byte, comma bool, d2 byte, c byte) []byte {
	rgba := image.NewRGBA(image.Rect(0, 0, 16, 16))

	g := glyphs[d1]
	i0 := 1
	j0 := 4
	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			if g[j]&(1<<(3-i)) > 0 {
				rgba.Set(i+i0, j+j0, color.Black)
			}
		}
	}
	if comma {
		rgba.Set(6, 12, color.Black)
		rgba.Set(6, 13, color.Black)
		rgba.Set(5, 14, color.Black)
	}
	g = glyphs[d2]
	i0 = 6
	j0 = 4
	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			if g[j]&(1<<(3-i)) > 0 {
				rgba.Set(i+i0, j+j0, color.Black)
			}
		}
	}
	g = glyphs[c]
	i0 = 11
	j0 = 4
	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			if g[j]&(1<<(3-i)) > 0 {
				rgba.Set(i+i0, j+j0, color.Black)
			}
		}
	}
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
