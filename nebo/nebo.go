package nebo

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/matematik7/jaslice-go/application"
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

	if strings.HasPrefix(command, "mode/") {
		mode, err := strconv.Atoi(command[5:])
		if err != nil {
			log.Println("Error decoding mode:", err)
			w.WriteHeader(500)
			return
		}

		if mode >= len(modes) || mode < 0 {
			log.Println("Mode out of range")
			w.WriteHeader(500)
			return
		}

		nebo.setMode(byte(mode))
	} else if strings.HasPrefix(command, "speed/") {
		speed, err := strconv.Atoi(command[6:])
		if err != nil {
			log.Println("Error decoding speed:", err)
			w.WriteHeader(500)
			return
		}

		if speed > 255 || speed < 0 {
			log.Println("Speed out of range")
			w.WriteHeader(500)
			return
		}

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
