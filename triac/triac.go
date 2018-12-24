package triac

import (
	"fmt"
	"log"
	"net/http"

	"github.com/matematik7/jaslice-go/application"
)

const address = 0x53

type Triac struct {
	app *application.App

	on    bool
	value byte
	relay byte

	currentOn    bool
	currentValue byte
}

func New(app *application.App, config application.Config) application.Module {
	return &Triac{
		app:   app,
		on:    config.GetBool("on"),
		value: config.GetByte("value"),
		relay: config.GetByte("relay"),
	}
}

func (triac *Triac) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Path

	if command == "on" {
		if !triac.currentOn {
			triac.turnOn()
		}
	} else if command == "off" {
		if triac.currentOn {
			triac.turnOff()
		}
	} else if command == "toggle" {
		if triac.currentOn {
			triac.turnOff()
		} else {
			triac.turnOn()
		}
	} else if value, ok := application.CommandInt(w, command, "value/", 0, 255); ok {
		triac.setValue(byte(value))
	} else {
		w.WriteHeader(404)
	}
}

type data struct {
	On    bool
	Value byte
}

func (triac *Triac) Data() interface{} {
	return data{
		On:    triac.currentOn,
		Value: triac.currentValue,
	}
}

func (triac *Triac) turnOn() {
	triac.changeRelay("on")
	triac.app.I2cBus.WriteByte(address, triac.currentValue)
	triac.currentOn = true
}

func (triac *Triac) turnOff() {
	triac.app.I2cBus.WriteByte(address, 0)
	triac.changeRelay("off")
	triac.currentOn = false
}

func (triac *Triac) changeRelay(state string) {
	url := fmt.Sprintf("http://localhost/api/relay%d/%s", triac.relay, state)
	_, err := http.Get(url)
	if err != nil {
		log.Printf("Could not set relay %d to %s: %v", triac.relay, state, err)
	}

}

func (triac *Triac) setValue(v byte) {
	triac.currentValue = v
	triac.app.I2cBus.WriteByte(address, triac.currentValue)
}

func (triac *Triac) On() {
	triac.currentValue = triac.value
	if triac.on {
		triac.turnOn()
	}
}

func (triac *Triac) Off() {
	triac.turnOff()
}
