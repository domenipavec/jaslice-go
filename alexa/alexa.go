package alexa

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

	"github.com/fromkeith/gossdp"
	"github.com/matematik7/jaslice-go/application"
	uuid "github.com/satori/go.uuid"
)

const SetupXML = `<?xml version="1.0"?>
<root>
  <device>
    <deviceType>urn:MakerMusings:device:controllee:1</deviceType>
    <friendlyName>%s</friendlyName>
    <manufacturer>Belkin International Inc.</manufacturer>
    <modelName>Emulated Socket</modelName>
    <modelNumber>3.1415</modelNumber>
    <UDN>uuid:%s</UDN>
  </device>
</root>
`

type Alexa struct {
	invocation string
	onUrl      string
	offUrl     string
}

func New(app *application.App, config application.Config) application.Module {
	alexa := &Alexa{
		invocation: config.GetString("invocation"),
		onUrl:      config.GetString("onUrl"),
		offUrl:     config.GetString("offUrl"),
	}

	port := 40000 + rand.Intn(10000)
	serial := "Socket-1_0-" + uuid.NewV4().String()

	serverDef := gossdp.AdvertisableServer{
		ServiceType: "urn:Belkin:device:**",
		DeviceUuid:  serial,
		Location:    fmt.Sprintf("http://172.23.163.16:%d/setup.xml", port),
		MaxAge:      86400,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/setup.xml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, SetupXML, alexa.invocation, serial)
	})
	mux.HandleFunc("/upnp/control/basicevent1", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading body:", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
		}

		action := ""
		if bytes.Contains(body, []byte("<BinaryState>1</BinaryState>")) {
			action = "on"
		} else if bytes.Contains(body, []byte("<BinaryState>0</BinaryState>")) {
			action = "off"
		} else {
			log.Println("Invalid action")
			return
		}

		log.Printf("got control event for %s: %s", alexa.invocation, action)
		if action == "on" {
			alexa.get(alexa.onUrl)
		} else {
			alexa.get(alexa.offUrl)
		}
	})

	go http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	app.Ssdp.AdvertiseServer(serverDef)

	return alexa
}

func (alexa *Alexa) get(url string) {
	_, err := http.Get("http://localhost:80" + url)
	if err != nil {
		log.Printf("Error with get request to %s: %s", url, err)
	}
}

func (alexa *Alexa) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}

func (alexa *Alexa) Data() interface{} {
	return nil
}

func (alexa *Alexa) On() {

}

func (alexa *Alexa) Off() {

}
