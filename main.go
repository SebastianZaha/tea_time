package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/SebastianZaha/systray"
	"github.com/SebastianZaha/webview"
)

//go:embed ui/index.html
var templateFS embed.FS

//go:embed ui/*.css
var ui embed.FS

var addr = "localhost:3009"
var srv = &http.Server{}

var templates *template.Template

var win *window

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

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	uiQuit := make(chan bool)

	go func() {
		select {
		case <-sig:
		case <-uiQuit:
		}

		// Shutdown signal with grace period
		shutdownCtx, _ := context.WithTimeout(serverCtx, 5*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()
	go func() { log.Fatal(srv.Serve(listener)) }()
	go func() {
		win = NewWindow()

		win.Run()
		uiQuit <- true
	}()

	// Wait for server context to be stopped
	<-serverCtx.Done()
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
				win.Show()
			}()

			win.Close()
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

type window struct {
	view webview.WebView
}

func NewWindow() *window {
	w := &window{view: webview.New(true)}
	w.view.SetTitle("TeaTime")
	w.view.SetSize(680, 480, webview.HintNone)
	w.view.Navigate("http://" + addr)
	systray.Register(onReady(w.view))
	return w
}

func (w *window) Run() {
	log.Println("running")
	w.view.Run() // blocks
	log.Println("destroying")
	w.view.Destroy()
	w.view = nil
}

func (w *window) Show() {
	w.view.Show()
}

func (w *window) Close() {
	w.view.Hide()
}

func onReady(w webview.WebView) func() {
	return func() {
		systray.SetTitle("TeaTime")
		mQuit := systray.AddMenuItem("Quit", "Quit")

		go func() {
			for {
				select {
				case <-mQuit.ClickedCh:
					w.Terminate()
				}
			}
		}()
	}
}
