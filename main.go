package main

import (
	"log"

	"github.com/Zemanta/gracefulshutdown"
	"github.com/Zemanta/gracefulshutdown/shutdownmanagers/posixsignal"

	"github.com/matematik7/jaslice-go/application"
	"github.com/matematik7/jaslice-go/fire"
	"github.com/matematik7/jaslice-go/musicplayer"
)

func main() {
	gs := gracefulshutdown.New()
	gs.AddShutdownManager(posixsignal.NewPosixSignalManager())
	gs.SetErrorHandler(gracefulshutdown.ErrorFunc(func(err error) {
		log.Println("Error grafefulshutdown:", err)
	}))

	app := application.New()

	app.AddModule("musicplayer", musicplayer.New)
	app.AddModule("fire", fire.New)

	app.Initialize("config.json")

	gs.AddShutdownCallback(app)
	if err := gs.Start(); err != nil {
		log.Println("Error starting gracefulshutdown:", err)
	}

	app.Start()
}
