package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/SebastianZaha/webview"
)

//go:embed ui/index.html
var templateFS embed.FS

//go:embed ui/*.css
var ui embed.FS

var addr = "localhost:3009"
var srv = &http.Server{}

var view webview.WebView
var templates *template.Template

func main() {
	// listen first to be sure that the webview does not error out when fetching
	// initial page, since we are serving the site in a goroutine
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

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
	srv := &http.Server{Handler: router}
	go func() { log.Fatal(srv.Serve(listener)) }()

	view := webview.New(true)
	defer view.Destroy()
	view.SetTitle("TeaTime")
	view.SetSize(680, 480, webview.HintNone)
	view.Navigate("http://" + addr)
	view.Run()
}

type Response struct {
	Error string
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
				for {
					select {
					case <-done:
						return
					case t := <-ticker.C:
						fmt.Println("Tick at", t)
					}
				}
			}()
			timer := time.NewTimer(duration)
			go func() {
				<-timer.C
				ticker.Stop()
				done <- true
			}()

		}
	}

	render(response, w)
}

func render(response Response, w http.ResponseWriter) {
	err := templates.ExecuteTemplate(w, "index.html", response)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error displaying page", http.StatusInternalServerError)
		return
	}
}
