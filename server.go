package main

import (
	"embed"
	"errors"
	"html/template"
	"image/color"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/SebastianZaha/systray"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

//go:embed ui/index.html
var templateFS embed.FS

//go:embed ui/*.css
var ui embed.FS

var addr = "localhost:3009"
var srv = &http.Server{}

var templates *template.Template

var colorHour = "#9afcba"
var colorMin = "#faefa3"
var colorSec = "#ff978a"

type Response struct {
	Error string
}

func newHandler() http.Handler {
	var err error
	templates, err = template.ParseFS(templateFS, "ui/*.html")
	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(10 * time.Second))

	router.Handle("/ui/*", http.StripPrefix("/", http.FileServer(http.FS(ui))))

	router.Get("/", handle)

	log.Printf("Running server on %s", addr)
	return router
}

func handle(w http.ResponseWriter, r *http.Request) {
	response := Response{}

	q := r.FormValue("q")
	if q != "" {
		duration, err := time.ParseDuration(q)
		if err != nil {
			response.Error = err.Error()
		} else {
			ticker := time.NewTicker(time.Second)
			done := make(chan bool)

			go func() {
				i := 0
				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						var icon []byte
						i++
						left := duration - time.Duration(i)*time.Second
						if left > 99*time.Hour {
							c, err := ParseHexColor(colorHour)
							if err != nil {
								log.Fatal(err)
							}
							icon = drawText("++", c)
						} else if left > time.Hour {
							c, err := ParseHexColor(colorHour)
							if err != nil {
								log.Fatal(err)
							}

							icon = drawText(strconv.Itoa(int(left/time.Hour)), c)
						} else if left > 99*time.Second {
							c, err := ParseHexColor(colorMin)
							if err != nil {
								log.Fatal(err)
							}

							icon = drawText(strconv.Itoa(int(left/time.Minute)), c)
						} else {
							c, err := ParseHexColor(colorSec)
							if err != nil {
								log.Fatal(err)
							}

							icon = drawText(strconv.Itoa(int(left/time.Second)), c)
						}
						systray.SetTemplateIcon(icon, icon)
					}
				}
			}()
			timer := time.NewTimer(duration)
			go func() {
				<-timer.C
				ticker.Stop()
				done <- true
				win.Show()
			}()

			win.Close()
		}
	}

	renderIndex(response, w)
}

func renderIndex(response Response, w http.ResponseWriter) {
	err := templates.ExecuteTemplate(w, "index.html", response)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error displaying page", http.StatusInternalServerError)
		return
	}
}

// https://stackoverflow.com/a/54200713/1306453
var errInvalidFormat = errors.New("invalid format")

func ParseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff

	if s[0] != '#' {
		return c, errInvalidFormat
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
		err = errInvalidFormat
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
		err = errInvalidFormat
	}
	return
}
