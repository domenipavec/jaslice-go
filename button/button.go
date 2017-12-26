package button

import (
	"log"
	"net/http"
	"time"

	"github.com/matematik7/jaslice-go/application"
	"github.com/stianeikeland/go-rpio"
)

const beforePoll = time.Second * 5
const poll = time.Millisecond * 25
const shortN = 2
const longN = 60

type Button struct {
	pinName string

	shortUrl string
	longUrl  string

	pin rpio.Pin
}

func New(app *application.App, config application.Config) application.Module {
	button := &Button{
		pin:      rpio.Pin(config.GetInt("pin")),
		shortUrl: config.GetString("shortUrl"),
		longUrl:  config.GetString("longUrl"),
	}

	if app.GpioEnabled {
		go button.poll()
	}

	return button
}

func (button *Button) poll() {
	time.Sleep(beforePoll)

	button.pin.Input()
	button.pin.PullUp()

	i := 0
	for {
		value := button.pin.Read()

		if value == rpio.High {
			if i != 0 {
				if i >= shortN && i < longN {
					button.get(button.shortUrl)
				}
				i = 0
			}
		} else {
			if i == longN {
				button.get(button.longUrl)
			}
			i++
		}

		time.Sleep(poll)
	}

}

func (button *Button) get(url string) {
	_, err := http.Get("http://localhost:80" + url)
	if err != nil {
		log.Printf("Error with get request to %s: %s", url, err)
	}
}

func (button *Button) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}

func (button *Button) Data() interface{} {
	return nil
}

func (button *Button) On() {

}

func (button *Button) Off() {

}
