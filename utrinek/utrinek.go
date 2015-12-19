package utrinek

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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

func New(app *application.App, config map[string]interface{}) application.Module {
	utrinek := &Utrinek{
		app:     app,
		onChan:  make(chan bool),
		minChan: make(chan int),
		maxChan: make(chan int),
	}

	go utrinek.worker()

	address, success := config["address"].(float64)
	if !success {
		log.Fatalln("Unable to parse utrinek address:", config)
	}
	utrinek.address = byte(address)

	min, success := config["min"].(float64)
	if !success {
		log.Fatalln("Unable to parse utrinek min:", config)
	}
	utrinek.min = int(min)

	max, success := config["max"].(float64)
	if !success {
		log.Fatalln("Unable to parse utrinek max:", config)
	}
	utrinek.max = int(max)

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
	} else if strings.HasPrefix(command, "min/") {
		min, err := strconv.Atoi(command[4:])
		if err != nil {
			log.Println("Error decoding min:", err)
			w.WriteHeader(500)
			return
		}

		if min < 0 || min > utrinek.currentMax {
			log.Println("Utrinek min out of range")
			w.WriteHeader(500)
			return
		}

		utrinek.minChan <- min
	} else if strings.HasPrefix(command, "max/") {
		max, err := strconv.Atoi(command[4:])
		if err != nil {
			log.Println("Error decoding max:", err)
			w.WriteHeader(500)
			return
		}

		if max < utrinek.currentMin {
			log.Println("Light out of range")
			w.WriteHeader(500)
			return
		}

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
	seconds := utrinek.currentMin + rand.Intn(diff)
	duration := time.Duration(seconds) * time.Second
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
