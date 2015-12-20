package application

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	rpio "github.com/stianeikeland/go-rpio"
)

const waitForPower = time.Millisecond * 100
const waitBetweenInits = time.Millisecond * 10

type Module interface {
	http.Handler

	Data() interface{}
	On()
	Off()
}

type ModuleConstructor func(*App, Config) Module

type ModuleData struct {
	Name      string
	ModuleId  string
	Template  string
	UrlPrefix string

	Module Module
}

type ModuleConfig struct {
	Module string `json:"module"`
	Name   string `json:"name"`
	Config Config `json:"config"`
}

type App struct {
	template *template.Template

	Modules            []ModuleData
	moduleConstructors map[string]ModuleConstructor
	moduleCounts       map[string]int

	On       bool
	powerPin rpio.Pin

	I2cBus embd.I2CBus
}

func New() *App {
	app := &App{
		moduleCounts:       make(map[string]int),
		moduleConstructors: make(map[string]ModuleConstructor),
		template:           template.New("index.html"),
	}

	app.template.Funcs(map[string]interface{}{
		"ModuleTemplate": func(name string, data interface{}) (ret template.HTML, err error) {
			buf := bytes.NewBuffer([]byte{})
			err = app.template.ExecuteTemplate(buf, name, data)
			ret = template.HTML(buf.String())
			return
		},
	})

	return app
}

func (app *App) templateName(name string) string {
	return name + ".html"
}

func (app *App) templateFile(name string) string {
	fn := "./templates/" + app.templateName(name)
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		return ""
	}
	return fn
}

func (app *App) AddModule(name string, mc ModuleConstructor) {
	app.moduleConstructors[name] = mc

	if fn := app.templateFile(name); fn != "" {
		if _, err := app.template.ParseFiles(fn); err != nil {
			log.Fatalln("Error parsing template", name, ":", err)
		}
	}

	log.Println("Added module:", name)
}

func (app *App) Initialize(configFn string) {
	if _, err := app.template.ParseFiles(app.templateFile("index")); err != nil {
		log.Fatalln("Error parsing index:", err)
	}

	configFile, err := os.Open(configFn)
	if err != nil {
		log.Fatalln("Error opening config:", err)
	}
	defer configFile.Close()

	config := []ModuleConfig{}

	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&config); err != nil {
		log.Fatalln("Error decoding config:", err)
	}

	for _, moduleConfig := range config {
		app.initModule(moduleConfig)
	}
}

func (app *App) initModule(moduleConfig ModuleConfig) {
	module := app.moduleConstructors[moduleConfig.Module](app, moduleConfig.Config)

	app.moduleCounts[moduleConfig.Module] += 1
	instanceNumber := strconv.Itoa(app.moduleCounts[moduleConfig.Module])

	urlPrefix := "/api/" + moduleConfig.Module + instanceNumber + "/"

	template := ""
	if app.templateFile(moduleConfig.Module) != "" {
		template = app.templateName(moduleConfig.Module)
	}

	app.Modules = append(app.Modules, ModuleData{
		Name:      moduleConfig.Name,
		ModuleId:  moduleConfig.Module,
		UrlPrefix: urlPrefix,
		Template:  template,
		Module:    module,
	})

	http.Handle(urlPrefix, http.StripPrefix(urlPrefix, module))

	log.Println("Initialized module:", moduleConfig.Name)
}

func (app *App) Start() {
	var err error
	time.Sleep(time.Second)

	if err = rpio.Open(); err != nil {
		log.Fatalln("Error init gpio:", err)
	}

	app.I2cBus = embd.NewI2CBus(1)

	app.powerPin = rpio.Pin(18)
	app.powerPin.Mode(rpio.Output)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/", app)

	log.Fatalln("Error serve:", http.ListenAndServe(":80", nil))
}

func (app *App) OnShutdown(string) error {
	app.turnOff()

	app.I2cBus.Close()
	rpio.Close()

	return nil
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		err := app.template.ExecuteTemplate(w, "index.html", app)
		if err != nil {
			log.Println("Error executing template:", err)
			w.WriteHeader(500)
		}
	} else if r.URL.Path == "/api/on" {
		if !app.On {
			app.turnOn()
		}
	} else if r.URL.Path == "/api/off" {
		if app.On {
			app.turnOff()
		}
	} else if r.URL.Path == "/api/toggle" {
		if app.On {
			app.turnOff()
		} else {
			app.turnOn()
		}
	} else {
		w.WriteHeader(404)
	}
}

func (app *App) turnOn() {
	app.On = true

	app.powerPin.Write(rpio.High)

	time.Sleep(waitForPower)

	for _, module := range app.Modules {
		module.Module.On()
		time.Sleep(waitBetweenInits)
	}
}

func (app *App) turnOff() {
	app.On = false

	for _, module := range app.Modules {
		module.Module.Off()
	}

	app.powerPin.Write(rpio.Low)
}
