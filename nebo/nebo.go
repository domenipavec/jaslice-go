package nebo

import (
	"net/http"

	"github.com/domenipavec/jaslice-go/application"
)

const address = 0x50

var modes = []string{"Ugasnjeno", "Normalno", "Ozvezdja", "Enakomerno", "Utripanje posamezno", "Utripanje več"}

type Nebo struct {
	app *application.App

	mode  byte
	speed byte
}

func New(app *application.App, config application.Config) application.Module {
	return &Nebo{
		app: app,
	}
}

func (nebo *Nebo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Path

	if mode, ok := application.CommandInt(w, command, "mode/", 0, len(modes)-1); ok {
		nebo.setMode(byte(mode))
	} else if speed, ok := application.CommandInt(w, command, "speed/", 0, 255); ok {
		nebo.setSpeed(byte(speed))
	} else {
		w.WriteHeader(404)
	}
}

type data struct {
	Modes []string
	Mode  byte
	Speed byte
}

func (nebo *Nebo) Data() interface{} {
	return data{
		Modes: modes,
		Mode:  nebo.mode,
		Speed: nebo.speed,
	}
}

func (nebo *Nebo) setMode(mode byte) {
	nebo.mode = mode
	nebo.app.I2cBus.WriteByteToReg(address, 0, mode)
}

func (nebo *Nebo) setSpeed(speed byte) {
	nebo.speed = speed
	nebo.app.I2cBus.WriteByteToReg(address, 1, speed)
}

func (nebo *Nebo) On() {
	nebo.setMode(1)
	nebo.setSpeed(19)
}

func (nebo *Nebo) Off() {
	nebo.setMode(0)
}
