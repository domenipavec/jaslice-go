package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/Zemanta/gracefulshutdown"
	"github.com/Zemanta/gracefulshutdown/shutdownmanagers/posixsignal"

	"github.com/domenipavec/jaslice-go/alexa"
	"github.com/domenipavec/jaslice-go/application"
	"github.com/domenipavec/jaslice-go/button"
	"github.com/domenipavec/jaslice-go/fire"
	"github.com/domenipavec/jaslice-go/luna"
	"github.com/domenipavec/jaslice-go/musicplayer"
	"github.com/domenipavec/jaslice-go/nebo"
	"github.com/domenipavec/jaslice-go/pwm"
	"github.com/domenipavec/jaslice-go/relay"
	"github.com/domenipavec/jaslice-go/triac"
	"github.com/domenipavec/jaslice-go/utrinek"
	"github.com/domenipavec/jaslice-go/utrinki"
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
	app.AddModule("utrinki", utrinki.New)
	app.AddModule("luna", luna.New)
	app.AddModule("button", button.New)
	app.AddModule("alexa", alexa.New)
	app.AddModule("triac", triac.New)

	app.Initialize("config.json")

	gs.AddShutdownCallback(app)
	if err := gs.Start(); err != nil {
		log.Println("Error starting gracefulshutdown:", err)
	}

	app.Start()
}
