package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"
)

var uiQuit = make(chan bool)

func main() {
	systray.Run(onReady, func() {})
}

func onReady() {
	systray.SetTitle("TeaTime")
	if icon := iconInactive(); icon != nil {
		systray.SetTemplateIcon(icon, icon)
	}
	mQuit := systray.AddMenuItem("Quit", "Quit")

	if len(os.Args) == 2 {
		onTimerStart(os.Args[1])
	} else {
		q, err := zenity.Entry("Enter a duration like 45s or 1.5m or 2h2m2s or 10",
			zenity.Title("Enter duration"))
		if err != nil {
			os.Exit(0)
		}
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
			if icon := iconForDuration(duration); icon != nil {
				systray.SetTemplateIcon(icon, icon)
			}
			systray.SetTitle(duration.String())
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
						if icon := iconForDuration(left); icon != nil {
							systray.SetTemplateIcon(icon, icon)
						}
						systray.SetTitle(left.String())
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

func onTimerDone(q string) {
	zenity.Info(fmt.Sprintf("Timer complete (%s).", q),
		zenity.Title("Done"),
		zenity.InfoIcon)
	systray.Quit()
}
