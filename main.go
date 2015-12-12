package main

import (
	"github.com/matematik7/jaslice-go/application"
	"github.com/matematik7/jaslice-go/musicplayer"
)

func main() {
	app := application.New()

	app.AddModule("musicplayer", musicplayer.New)

	app.Initialize("config.json")

	app.Start()
}
