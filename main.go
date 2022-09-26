package main

import (
	"fmt"
	"log"
	"math"
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
	icon := draw(13, false, 3, 14)
	systray.SetTemplateIcon(icon, icon)
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
	zenity.Info(fmt.Sprintf("Timer complete (%s).", q),
		zenity.Title("Done"),
		zenity.InfoIcon)
	systray.Quit()
}
