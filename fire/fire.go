package fire

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/matematik7/jaslice-go/application"
)

type Fire struct {
	app *application.App

	address byte
	speed   byte
	color   byte
	light   byte

	currentSpeed byte
	currentColor byte
	currentLight byte

	on bool
}

func New(app *application.App, config application.Config) application.Module {
	return &Fire{
		app:     app,
		address: config.GetByte("address"),
		speed:   config.GetByte("speed"),
		color:   config.GetByte("color"),
		light:   config.GetByte("light"),
	}
}

func (fire *Fire) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Path

	if command == "on" {
		fire.On()
	} else if command == "off" {
		fire.Off()
	} else if strings.HasPrefix(command, "color/") {
		color, err := strconv.Atoi(command[6:])
		if err != nil {
			log.Println("Error decoding color:", err)
			w.WriteHeader(500)
			return
		}

		if color > 255 || color < 0 {
			log.Println("Color out of range")
			w.WriteHeader(500)
			return
		}

		fire.setColor(byte(color))
	} else if strings.HasPrefix(command, "light/") {
		light, err := strconv.Atoi(command[6:])
		if err != nil {
			log.Println("Error decoding light:", err)
			w.WriteHeader(500)
			return
		}

		if light > 255 || light < 0 {
			log.Println("Light out of range")
			w.WriteHeader(500)
			return
		}

		fire.setLight(byte(light))
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

		fire.setSpeed(byte(speed))
	} else {
		w.WriteHeader(404)
	}
}

type data struct {
	On bool

	Color byte
	Light byte
	Speed byte
}

func (fire *Fire) Data() interface{} {
	return data{
		On:    fire.on,
		Color: fire.currentColor,
		Light: fire.currentLight,
		Speed: fire.currentSpeed,
	}
}

func (fire *Fire) setColor(color byte) {
	fire.currentColor = color
	fire.app.I2cBus.WriteByteToReg(fire.address, 2, color)
}

func (fire *Fire) setLight(light byte) {
	fire.currentLight = light
	fire.app.I2cBus.WriteByteToReg(fire.address, 3, light)
}

func (fire *Fire) setSpeed(speed byte) {
	fire.currentSpeed = speed
	fire.app.I2cBus.WriteByteToReg(fire.address, 4, speed)
}

func (fire *Fire) On() {
	fire.on = true
	fire.app.I2cBus.WriteByte(fire.address, 1)
	fire.setColor(fire.color)
	fire.setLight(fire.light)
	fire.setSpeed(fire.speed)
}

func (fire *Fire) Off() {
	fire.on = false
	fire.app.I2cBus.WriteByte(fire.address, 0)
}
