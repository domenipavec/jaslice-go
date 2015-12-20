package application

import (
	"log"

	_ "github.com/kidoman/embd/host/rpi"
)

type Config map[string]interface{}

func (c Config) Get(name string) interface{} {
	value, ok := c[name]
	if !ok {
		log.Fatalln("Config value not defined:", name)
	}
	return value
}

func (c Config) GetByte(name string) byte {
	value, ok := c.Get(name).(float64)
	if !ok {
		log.Fatalf("Config value %s is not proper type: %T", name, c.Get(name))
	}
	return byte(value)
}

func (c Config) GetBool(name string) bool {
	value, ok := c.Get(name).(bool)
	if !ok {
		log.Fatalf("Config value %s is not proper type: %T", name, c.Get(name))
	}
	return value
}

func (c Config) GetInt(name string) int {
	value, ok := c.Get(name).(float64)
	if !ok {
		log.Fatalf("Config value %s is not proper type: %T", name, c.Get(name))
	}
	return int(value)
}

func (c Config) GetString(name string) string {
	value, ok := c.Get(name).(string)
	if !ok {
		log.Fatalf("Config value %s is not proper type: %T", name, c.Get(name))
	}
	return value
}

func (c Config) GetSlice(name string) []interface{} {
	value, ok := c.Get(name).([]interface{})
	if !ok {
		log.Fatalf("Config value %s is not a slice, is: %T", name, c.Get(name))
	}
	return value
}

func (c Config) GetSliceStrings(name string) []string {
	data := []string{}
	for _, item := range c.GetSlice(name) {
		value, ok := item.(string)
		if !ok {
			log.Fatalln("List item is not string:", name)
		}
		data = append(data, value)
	}
	return data
}
