package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/Zemanta/gracefulshutdown"
	"github.com/Zemanta/gracefulshutdown/shutdownmanagers/posixsignal"

	"github.com/matematik7/jaslice-go/application"
	"github.com/matematik7/jaslice-go/button"
	"github.com/matematik7/jaslice-go/fire"
	"github.com/matematik7/jaslice-go/luna"
	"github.com/matematik7/jaslice-go/musicplayer"
	"github.com/matematik7/jaslice-go/nebo"
	"github.com/matematik7/jaslice-go/pwm"
	"github.com/matematik7/jaslice-go/relay"
	"github.com/matematik7/jaslice-go/utrinek"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	gs := gracefulshutdown.New()
	gs.AddShutdownManager(posixsignal.NewPosixSignalManager())
	gs.SetErrorHandler(gracefulshutdown.ErrorFunc(func(err error) {
		log.Println("Error grafefulshutdown:", err)
	}))

	app := application.New()

	app.AddModule("musicplayer", musicplayer.New)
	app.AddModule("fire", fire.New)
	app.AddModule("nebo", nebo.New)
	app.AddModule("pwm", pwm.New)
	app.AddModule("relay", relay.New)
	app.AddModule("utrinek", utrinek.New)
	app.AddModule("button", button.New)
	app.AddModule("luna", luna.New)

	app.Initialize("config.json")

	gs.AddShutdownCallback(app)
	if err := gs.Start(); err != nil {
		log.Println("Error starting gracefulshutdown:", err)
	}

	app.Start()
}
