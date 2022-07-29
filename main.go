package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"
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
	icon := drawText("Tea", color.White)
	systray.SetTemplateIcon(icon, icon)
	mQuit := systray.AddMenuItem("Quit", "Quit")

	if len(os.Args) == 2 {
		onTimerStart(os.Args[1])
	} else {
		q, err := zenity.Entry("Enter duration (e.g. 45s or 25m or 2h2m2s):",
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
	if d > 99*time.Hour {
		return drawText("++", colorHour)
	} else if d > time.Hour {
		return drawText(strconv.Itoa(int(d/time.Hour)), colorHour)
	} else if d > 99*time.Second {
		return drawText(strconv.Itoa(int(d/time.Minute)), colorMin)
	} else {
		return drawText(strconv.Itoa(int(d/time.Second)), colorSec)
	}
}

func onTimerDone(q string) {
	zenity.Info(fmt.Sprintf("Timer complete (%s).", q),
		zenity.Title("Done"),
		zenity.InfoIcon)
	systray.Quit()
}
