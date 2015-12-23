package utrinek

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/matematik7/jaslice-go/application"
)

type Utrinek struct {
	app *application.App

	address byte
	min     int
	max     int

	currentMin int
	currentMax int

	on bool

	onChan  chan bool
	minChan chan int
	maxChan chan int
}

func New(app *application.App, config application.Config) application.Module {
	utrinek := &Utrinek{
		app:     app,
		address: config.GetByte("address"),
		min:     config.GetInt("min"),
		max:     config.GetInt("max"),
		onChan:  make(chan bool),
		minChan: make(chan int),
		maxChan: make(chan int),
	}

	go utrinek.worker()

	return utrinek
}

func (utrinek *Utrinek) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Path

	if command == "show" {
		utrinek.show()
	} else if command == "on" {
		utrinek.On()
	} else if command == "off" {
		utrinek.Off()
	} else if min, ok := application.CommandInt(w, command, "min/", 0, utrinek.currentMax); ok {
		utrinek.minChan <- min
	} else if max, ok := application.CommandInt(w, command, "max/", utrinek.currentMin, 100000); ok {
		utrinek.maxChan <- max
	} else {
		w.WriteHeader(404)
	}
}

type data struct {
	On bool

	Min int
	Max int
}

func (utrinek *Utrinek) Data() interface{} {
	return data{
		On:  utrinek.on,
		Min: utrinek.currentMin,
		Max: utrinek.currentMax,
	}
}

func (utrinek *Utrinek) show() {
	utrinek.app.I2cBus.WriteByte(utrinek.address, 0)
}

func (utrinek *Utrinek) getTimer() *time.Timer {
	diff := utrinek.currentMax - utrinek.currentMin
	if diff <= 0 {
		diff = 1
	}
	diff *= 1000
	seconds := 1000*utrinek.currentMin + rand.Intn(diff)
	duration := time.Duration(seconds) * time.Millisecond
	return time.NewTimer(duration)
}

func (utrinek *Utrinek) worker() {
	on := false
	timer := utrinek.getTimer()
	timer.Stop()
	for {
		select {
		case v := <-utrinek.onChan:
			on = v
			if on {
				timer = utrinek.getTimer()
			} else {
				timer.Stop()
			}
		case min := <-utrinek.minChan:
			utrinek.currentMin = min
			timer = utrinek.getTimer()
		case max := <-utrinek.maxChan:
			utrinek.currentMin = max
			timer = utrinek.getTimer()
		case <-timer.C:
			utrinek.show()
			timer = utrinek.getTimer()
		}
	}
}

func (utrinek *Utrinek) On() {
	utrinek.on = true
	utrinek.currentMax = utrinek.max
	utrinek.currentMin = utrinek.min
	utrinek.onChan <- true
}

func (utrinek *Utrinek) Off() {
	utrinek.on = false
	utrinek.onChan <- false
}
