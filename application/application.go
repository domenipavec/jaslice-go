package application

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Module interface {
	http.Handler

	Data() interface{}
}

type ModuleConstructor func(map[string]interface{}) Module

type ModuleData struct {
	Name      string
	Template  string
	UrlPrefix string
	Module    Module
}

type ModuleConfig struct {
	Module string                 `json:"module"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

type App struct {
	Template           *template.Template
	ModuleCounts       map[string]int
	Modules            []ModuleData
	ModuleConstructors map[string]ModuleConstructor
}

func New() *App {
	app := &App{
		ModuleCounts:       make(map[string]int),
		ModuleConstructors: make(map[string]ModuleConstructor),
		Template:           template.New("index.html"),
	}

	app.Template.Funcs(map[string]interface{}{
		"ModuleTemplate": func(name string, data interface{}) (ret template.HTML, err error) {
			buf := bytes.NewBuffer([]byte{})
			err = app.Template.ExecuteTemplate(buf, name, data)
			ret = template.HTML(buf.String())
			return
		},
	})

	if _, err := app.Template.ParseFiles(app.templateFile("index")); err != nil {
		log.Fatal(err)
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
	app.ModuleConstructors[name] = mc

	if _, err := app.Template.ParseFiles(app.templateFile(name)); err != nil {
		log.Fatal(err)
	}

	log.Println("Added module:", name)
}

func (app *App) Initialize(configFn string) {
	configFile, err := os.Open(configFn)
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()

	config := []ModuleConfig{}

	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&config); err != nil {
		log.Fatal(err)
	}

	for _, moduleConfig := range config {
		app.InitModule(moduleConfig)
	}
}

func (app *App) InitModule(moduleConfig ModuleConfig) {
	module := app.ModuleConstructors[moduleConfig.Module](moduleConfig.Config)

	app.ModuleCounts[moduleConfig.Module] += 1
	instanceNumber := strconv.Itoa(app.ModuleCounts[moduleConfig.Module])

	urlPrefix := "/api/" + moduleConfig.Module + instanceNumber + "/"

	app.Modules = append(app.Modules, ModuleData{
		Name:      moduleConfig.Name,
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

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := app.Template.ExecuteTemplate(w, "index.html", app.Modules)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
	}
}
