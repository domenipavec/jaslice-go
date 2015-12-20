package relay

import (
	"net/http"

	"github.com/matematik7/jaslice-go/application"
)

const address = 0x51

type Relay struct {
	app *application.App

	on    bool
	index byte

	currentOn bool
}

func New(app *application.App, config application.Config) application.Module {
	return &Relay{
		app:   app,
		on:    config.GetBool("on"),
		index: config.GetByte("index"),
	}
}

func (relay *Relay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Path

	if command == "on" {
		if !relay.currentOn {
			relay.turnOn()
		}
	} else if command == "off" {
		if relay.currentOn {
			relay.turnOff()
		}
	} else if command == "toggle" {
		if relay.currentOn {
			relay.turnOff()
		} else {
			relay.turnOn()
		}
	} else {
		w.WriteHeader(404)
	}
}

type data struct {
	On bool
}

func (relay *Relay) Data() interface{} {
	return data{
		On: relay.currentOn,
	}
}

func (relay *Relay) turnOn() {
	relay.currentOn = true
	relay.app.I2cBus.WriteByteToReg(address, 1, relay.index)
}

func (relay *Relay) turnOff() {
	relay.currentOn = false
	relay.app.I2cBus.WriteByteToReg(address, 0, relay.index)
}

func (relay *Relay) On() {
	if relay.on {
		relay.turnOn()
	}
}

func (relay *Relay) Off() {
	relay.turnOff()
}
