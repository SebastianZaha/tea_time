package main

import (
	"bufio"
	"bytes"
	"fmt"
	ico "github.com/Kodeworks/golang-image-ico"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"
)

func main() {
	systray.Run(onReady, func() {})
}

func onReady() {
	icon := draw(13, false, 3, 14)
	systray.SetTemplateIcon(icon, icon)
	mQuit := systray.AddMenuItem("Quit", "Quit")

	if len(os.Args) == 2 {
		onTimerStart(os.Args[1])
	} else {
		msg := "Enter a duration like 45s or 1.5m or 2h2m2s or 10."
		hist := []string{}
		tmpfile := filepath.Join(os.TempDir(), "tea_time.txt")
		if _, err := os.Stat(tmpfile); err == nil {
			bs, _ := os.ReadFile(tmpfile)
			msg += " Prev timers: " + string(bs)
			hist = strings.Split(string(bs), " ")
		}
		q, err := zenity.Entry(msg, zenity.Title("Enter duration"))
		if err != nil {
			// either a problem with showing dialog, or cancel pressed
			os.Exit(0)
		}
		hist = append(hist, q)
		if len(hist) > 4 {
			hist = hist[1:]
		}
		_ = os.WriteFile(tmpfile, []byte(strings.Join(hist, " ")), 0666)
		onTimerStart(q)
	}

	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onTimerStart(q string) {
	if q != "" {
		// If it ends in a digit, assume seconds
		if '0' <= q[len(q)-1] && q[len(q)-1] <= '9' {
			q += "s"
		}

		duration, err := time.ParseDuration(q)
		if err != nil {
			log.Println(err)
		} else {
			icon := drawDuration(duration)
			systray.SetTemplateIcon(icon, icon)
			ticker := time.NewTicker(time.Second)
			done := make(chan bool)

			go func() {
				i := 0
				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						i++
						left := duration - time.Duration(i)*time.Second
						icon = drawDuration(left)
						systray.SetTemplateIcon(icon, icon)
					}
				}
			}()
			timer := time.NewTimer(duration)
			go func() {
				<-timer.C
				ticker.Stop()
				done <- true
				onTimerDone(q)
			}()
		}
	}
}

func drawDuration(d time.Duration) []byte {
	if d >= 100*time.Hour {
		return draw(11, false, 11, 11)
	} else if d >= 10*time.Hour {
		return draw(byte(d/time.Hour/10), false, byte((d/time.Hour)%10), 11)
	} else if d > time.Hour {
		i, frac := math.Modf(d.Hours())
		return draw(byte(i), true, byte(frac*10), 11)
	} else if d >= 10*time.Minute {
		return draw(byte(d/time.Minute/10), false, byte((d/time.Minute)%10), 10)
	} else if d > time.Minute {
		i, frac := math.Modf(d.Minutes())
		return draw(byte(i), true, byte(frac*10), 10)
	} else {
		return draw(byte(d/time.Second/10), false, byte(d/time.Second%10), 12)
	}
}

func onTimerDone(q string) {
	_ = zenity.Info(fmt.Sprintf("Timer complete (%s).", q),
		zenity.Title("Done"),
		zenity.InfoIcon)
	systray.Quit()
}

// font rendering

var glyphs = [][]byte{
	{0b0110, 0b1001, 0b1001, 0b1011, 0b1101, 0b1001, 0b1001, 0b0110}, // 0
	{0b0010, 0b0110, 0b0010, 0b0010, 0b0010, 0b0010, 0b0010, 0b0111}, // 1
	{0b0110, 0b1001, 0b0001, 0b0001, 0b0010, 0b0100, 0b1000, 0b1111}, // 2
	{0b1111, 0b0001, 0b0010, 0b0101, 0b0001, 0b0001, 0b1001, 0b0110}, // 3
	{0b1000, 0b1000, 0b1010, 0b1010, 0b1111, 0b0010, 0b0010, 0b0010},
	{0b1111, 0b1000, 0b1000, 0b1110, 0b0001, 0b0001, 0b1001, 0b0110}, // 5
	{0b0110, 0b1001, 0b1000, 0b1000, 0b1110, 0b1001, 0b1001, 0b0110},
	{0b1111, 0b0001, 0b0001, 0b0010, 0b0010, 0b0100, 0b0100, 0b0010}, // 7
	{0b0110, 0b1001, 0b1001, 0b0110, 0b0110, 0b1001, 0b1001, 0b0110},
	{0b0110, 0b1001, 0b1001, 0b0111, 0b0001, 0b0001, 0b1001, 0b0110}, // 9
	{0b0000, 0b0000, 0b0000, 0b1001, 0b1111, 0b1001, 0b1001, 0b1001}, // m 10
	{0b0000, 0b1000, 0b1000, 0b1110, 0b1001, 0b1001, 0b1001, 0b1001}, // h 11
	{0b0000, 0b0000, 0b0000, 0b0111, 0b1000, 0b0110, 0b0001, 0b1110}, // s 12
	{0b0000, 0b0100, 0b0100, 0b1111, 0b0100, 0b0100, 0b0100, 0b0011}, // t 13
	{0b0000, 0b0000, 0b0000, 0b0111, 0b1001, 0b1001, 0b1011, 0b0101}, // a 14
	{0b0000, 0b0000, 0b0000, 0b0000, 0b0000, 0b0000, 0b0000, 0b0000}, // space 15
}

func draw(d1 byte, comma bool, d2 byte, c byte) []byte {
	rgba := image.NewRGBA(image.Rect(0, 0, 16, 16))

	var txtColor color.Color
	if //goland:noinspection ALL
	runtime.GOOS == "windows" {
		txtColor = color.White
	} else {
		txtColor = color.Black
	}

	g := glyphs[d1]
	i0 := 1
	j0 := 4
	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			if g[j]&(1<<(3-i)) > 0 {
				rgba.Set(i+i0, j+j0, txtColor)
			}
		}
	}
	if comma {
		rgba.Set(6, 12, txtColor)
		rgba.Set(6, 13, txtColor)
		rgba.Set(5, 14, txtColor)
	}
	g = glyphs[d2]
	i0 = 6
	j0 = 4
	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			if g[j]&(1<<(3-i)) > 0 {
				rgba.Set(i+i0, j+j0, txtColor)
			}
		}
	}
	g = glyphs[c]
	i0 = 11
	j0 = 4
	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			if g[j]&(1<<(3-i)) > 0 {

				rgba.Set(i+i0, j+j0, txtColor)
			}
		}
	}
	var b bytes.Buffer
	var err error
	w := bufio.NewWriter(&b)

	if //goland:noinspection ALL
	runtime.GOOS == "windows" {
		err = ico.Encode(w, rgba)
	} else {
		err = png.Encode(w, rgba)
	}

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = w.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	bs, err := io.ReadAll(&b)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	return bs
}
