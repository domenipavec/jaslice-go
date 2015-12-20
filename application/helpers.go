package application

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

func CommandInt(w http.ResponseWriter, command, prefix string, min, max int) (int, bool) {
	if !strings.HasPrefix(command, prefix) {
		return 0, false
	}

	value, err := strconv.Atoi(command[len(prefix):])
	if err != nil {
		log.Printf("Error decoding %s: %s", prefix, err)
		w.WriteHeader(500)
		return 0, false
	}

	if value > max || value < min {
		log.Println(prefix, "out of range")
		w.WriteHeader(500)
		return 0, false
	}

	return value, true
}
