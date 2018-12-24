package utrinki

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/matematik7/jaslice-go/application"
)

const (
	TypeOld = "old"
	TypeNew = "new"
)

type Utrinki struct {
	app *application.App

	min int
	max int

	currentMin int
	currentMax int

	on bool

	onChan  chan bool
	minChan chan int
	maxChan chan int

	Configs []Utrinek
}

type Utrinek struct {
	Type    string
	Address byte
}

func New(app *application.App, config application.Config) application.Module {
	utrinki := &Utrinki{
		app:     app,
		min:     config.GetInt("min"),
		max:     config.GetInt("max"),
		onChan:  make(chan bool),
		minChan: make(chan int),
		maxChan: make(chan int),
	}

	for _, cfg := range config.GetSliceConfigs("utrinki") {
		utrinki.Configs = append(utrinki.Configs, Utrinek{
			Type:    cfg.GetString("type"),
			Address: byte(cfg.GetInt("address")),
		})
	}

	go utrinki.worker()

	return utrinki
}

func (utrinki *Utrinki) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Path

	if idx, ok := application.CommandInt(w, command, "show/", 0, len(utrinki.Configs)-1); ok {
		utrinki.show(idx)
	} else if idx, ok := application.CommandInt(w, command, "left/", 0, len(utrinki.Configs)-1); ok {
		utrinki.fromLeft(idx)
	} else if idx, ok := application.CommandInt(w, command, "right/", 0, len(utrinki.Configs)-1); ok {
		utrinki.fromRight(idx)
	} else if command == "on" {
		utrinki.On()
	} else if command == "off" {
		utrinki.Off()
	} else if min, ok := application.CommandInt(w, command, "min/", 0, utrinki.currentMax); ok {
		utrinki.minChan <- min
	} else if max, ok := application.CommandInt(w, command, "max/", utrinki.currentMin, 100000); ok {
		utrinki.maxChan <- max
	} else {
		w.WriteHeader(404)
	}
}

type data struct {
	On bool

	Min int
	Max int

	Configs []Utrinek
}

func (utrinki *Utrinki) Data() interface{} {
	return data{
		On:      utrinki.on,
		Min:     utrinki.currentMin,
		Max:     utrinki.currentMax,
		Configs: utrinki.Configs,
	}
}

func (utrinki *Utrinki) show(idx int) {
	cfg := utrinki.Configs[idx]
	if cfg.Type == TypeOld {
		if err := utrinki.app.I2cBus.WriteByte(cfg.Address, 0); err != nil {
			log.Printf("Error triggering old utrinek %d: %v", idx, err)
		}
	} else if cfg.Type == TypeNew {
		if rand.Intn(2) == 0 {
			utrinki.fromLeft(idx)
		} else {
			utrinki.fromRight(idx)
		}
	} else {
		log.Printf("Invalid config type %s for %d.", cfg.Type, idx)
	}
}

func (utrinki *Utrinki) fromLeft(idx int) {
	cfg := utrinki.Configs[idx]
	if cfg.Type != TypeNew {
		log.Printf("Cannot select from side for non new utrinek %d.", idx)
		return
	}

	utrinki.app.SerialPacket(cfg.Address, 0)
}

func (utrinki *Utrinki) fromRight(idx int) {
	cfg := utrinki.Configs[idx]
	if cfg.Type != TypeNew {
		log.Printf("Cannot select from side for non new utrinek %d.", idx)
		return
	}

	utrinki.app.SerialPacket(cfg.Address, 1)
}

func (utrinki *Utrinki) getTimer() *time.Timer {
	diff := utrinki.currentMax - utrinki.currentMin
	if diff <= 0 {
		diff = 1
	}
	diff *= 1000
	seconds := 1000*utrinki.currentMin + rand.Intn(diff)
	duration := time.Duration(seconds) * time.Millisecond
	return time.NewTimer(duration)
}

func (utrinki *Utrinki) worker() {
	on := false
	timer := utrinki.getTimer()
	timer.Stop()
	for {
		select {
		case v := <-utrinki.onChan:
			on = v
			if on {
				timer = utrinki.getTimer()
			} else {
				timer.Stop()
			}
		case min := <-utrinki.minChan:
			utrinki.currentMin = min
			timer = utrinki.getTimer()
		case max := <-utrinki.maxChan:
			utrinki.currentMax = max
			timer = utrinki.getTimer()
		case <-timer.C:
			utrinki.show(rand.Intn(len(utrinki.Configs)))
			timer = utrinki.getTimer()
		}
	}
}

func (utrinki *Utrinki) On() {
	utrinki.on = true
	utrinki.currentMax = utrinki.max
	utrinki.currentMin = utrinki.min
	utrinki.onChan <- true
}

func (utrinki *Utrinki) Off() {
	utrinki.on = false
	utrinki.onChan <- false
}
