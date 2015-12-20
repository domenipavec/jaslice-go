package luna

import (
	"net/http"

	"github.com/matematik7/jaslice-go/application"
)

const address = 0x52

var modes = []string{"Mlaj", "Prvi krajec 1", "Prvi krajec 2", "Prvi krajec 3", "Prvi krajec 4", "Polna luna", "Zadnji krajec 4", "Zadnji krajec 3", "Zadnji krajec 2", "Zadnji krajec 1"}

type Luna struct {
	app *application.App

	mode byte

	clientId     string
	clientSecret string
}

func New(app *application.App, config application.Config) application.Module {
	return &Luna{
		app:          app,
		clientId:     config.GetString("clientId"),
		clientSecret: config.GetString("clientSecret"),
	}
}

func (luna *Luna) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Path

	if mode, ok := application.CommandInt(w, command, "", 0, len(modes)-1); ok {
		luna.setMode(byte(mode))
	} else {
		w.WriteHeader(404)
	}
}

type data struct {
	Modes []string
	Mode  byte
}

func (luna *Luna) Data() interface{} {
	return data{
		Modes: modes,
		Mode:  luna.mode,
	}
}

func (luna *Luna) setMode(mode byte) {
	luna.mode = mode
	luna.app.I2cBus.WriteByte(address, mode)
}

func (luna *Luna) On() {
	luna.setMode(luna.getPhase())
}

func (luna *Luna) Off() {
	luna.setMode(0)
}
