package application

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

type Module interface {
	http.Handler

	Data() interface{}
	On()
	Off()
}

type ModuleConstructor func(map[string]interface{}) Module

type ModuleData struct {
	Name      string
	ModuleId  string
	Template  string
	UrlPrefix string

	Module Module
}

type ModuleConfig struct {
	Module string                 `json:"module"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

type App struct {
	template *template.Template

	Modules            []ModuleData
	moduleConstructors map[string]ModuleConstructor
	moduleCounts       map[string]int

	On       bool
	powerPin embd.DigitalPin
}

func New() *App {
	var err error

	if err := embd.InitGPIO(); err != nil {
		log.Fatalln("Error init gpio:", err)
	}

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

	if _, err = app.template.ParseFiles(app.templateFile("index")); err != nil {
		log.Fatalln("Error parsing index:", err)
	}

	app.powerPin, err = embd.NewDigitalPin("P1_12")
	if err != nil {
		log.Fatalln("Error creating power pin:", err)
	}

	err = app.powerPin.SetDirection(embd.Out)
	if err != nil {
		log.Fatalln("Error set direction for power pin:", err)
	}

	return app
}

func (app *App) templateName(name string) string {
	return name + ".html"
}

func (app *App) templateFile(name string) string {
	return "./templates/" + app.templateName(name)
}

func (app *App) AddModule(name string, mc ModuleConstructor) {
	app.moduleConstructors[name] = mc

	if _, err := app.template.ParseFiles(app.templateFile(name)); err != nil {
		log.Fatalln("Error parsing template", name, ":", err)
	}

	log.Println("Added module:", name)
}

func (app *App) Initialize(configFn string) {
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
		app.InitModule(moduleConfig)
	}
}

func (app *App) InitModule(moduleConfig ModuleConfig) {
	module := app.moduleConstructors[moduleConfig.Module](moduleConfig.Config)

	app.moduleCounts[moduleConfig.Module] += 1
	instanceNumber := strconv.Itoa(app.moduleCounts[moduleConfig.Module])

	urlPrefix := "/api/" + moduleConfig.Module + instanceNumber + "/"

	app.Modules = append(app.Modules, ModuleData{
		Name:      moduleConfig.Name,
		ModuleId:  moduleConfig.Module,
		Template:  app.templateName(moduleConfig.Module),
		UrlPrefix: urlPrefix,
		Module:    module,
	})

	http.Handle(urlPrefix, http.StripPrefix(urlPrefix, module))

	log.Println("Initialized module:", moduleConfig.Name)
}

func (app *App) Start() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/", app)

	log.Fatalln("Error serve:", http.ListenAndServe(":8080", nil))
}

func (app *App) OnShutdown(string) error {
	return embd.CloseGPIO()
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		err := app.template.ExecuteTemplate(w, "index.html", app)
		if err != nil {
			log.Println("Error executing template:", err)
			w.WriteHeader(500)
		}
	} else if r.URL.Path == "/api/on" {
		app.On = true

		if err := app.powerPin.Write(embd.High); err != nil {
			log.Println("Error writing power pin:", err)
			w.WriteHeader(500)
		}

		for _, module := range app.Modules {
			module.Module.On()
		}
	} else if r.URL.Path == "/api/off" {
		app.On = false

		for _, module := range app.Modules {
			module.Module.Off()
		}

		if err := app.powerPin.Write(embd.Low); err != nil {
			log.Println("Error writing power pin:", err)
			w.WriteHeader(500)
		}
	} else {
		w.WriteHeader(404)
	}
}
