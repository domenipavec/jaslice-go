package pwm

import (
	"log"
	"net/http"
	"strconv"

	"github.com/matematik7/jaslice-go/application"
)

const address = 0x50

type Pwm struct {
	app *application.App

	value byte
	index byte

	currentValue byte
}

func New(app *application.App, config application.Config) application.Module {
	return &Pwm{
		app:   app,
		value: config.GetByte("value"),
		index: config.GetByte("index"),
	}
}

func (pwm *Pwm) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	value, err := strconv.Atoi(r.URL.Path)
	if err != nil {
		log.Println("Error decoding value:", err)
		w.WriteHeader(500)
		return
	}

	if value > 255 || value < 0 {
		log.Println("Value out of range")
		w.WriteHeader(500)
		return
	}

	pwm.setValue(byte(value))
}

type data struct {
	Value byte
}

func (pwm *Pwm) Data() interface{} {
	return data{
		Value: pwm.currentValue,
	}
}

func (pwm *Pwm) setValue(value byte) {
	pwm.currentValue = value
	pwm.app.I2cBus.WriteByteToReg(address, 2+pwm.index, value)
}

func (pwm *Pwm) On() {
	pwm.setValue(pwm.value)
}

func (pwm *Pwm) Off() {

}
