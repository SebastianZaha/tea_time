package main

import (
	"context"
	"image/color"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SebastianZaha/systray"
	"github.com/SebastianZaha/webview"
)

var win *window

func main() {
	// listen first to be sure that the webview does not error out when fetching
	// initial page, since we are serving the site in a goroutine
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{Handler: newHandler()}

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

	win = NewWindow()
	win.Run()
	uiQuit <- true

	// Wait for server context to be stopped
	<-serverCtx.Done()
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
		icon := drawText("Tea", color.White)
		systray.SetTemplateIcon(icon, icon)
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
