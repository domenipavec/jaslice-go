package alexa

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/domenipavec/jaslice-go/application"
	"github.com/fromkeith/gossdp"
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

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func New(app *application.App, config application.Config) application.Module {
	alexa := &Alexa{
		invocation: config.GetString("invocation"),
		onUrl:      config.GetString("onUrl"),
		offUrl:     config.GetString("offUrl"),
	}

	uid := uuid.NewV3(uuid.NamespaceDNS, alexa.invocation)
	ip := GetOutboundIP()
	port := 40000 + int(uid[0]) + int(uid[1])*10
	serial := "Socket-1_0-" + uid.String()

	log.Printf("Running %s on %s:%d", serial, ip, port)

	serverDef := gossdp.AdvertisableServer{
		ServiceType: "urn:Belkin:device:**",
		DeviceUuid:  serial,
		Location:    fmt.Sprintf("http://%s:%d/setup.xml", ip, port),
		MaxAge:      86400,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/setup.xml", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("setup.xml for %s:%d", ip, port)
		fmt.Fprintf(w, SetupXML, alexa.invocation, serial)
	})
	mux.HandleFunc("/upnp/control/basicevent1", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading body:", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
		}

		if !bytes.Contains(body, []byte("SetBinaryState")) {
			return
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

	go func() {
		s := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		}
		s.SetKeepAlivesEnabled(false)
		err := s.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()
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
