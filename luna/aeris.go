package luna

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
)

const url = "http://api.aerisapi.com/sunmoon/ljubljana,slovenia?filter=moonphase&client_id=%s&client_secret=%s"

type phase struct {
	Age float64 `json:"age"`
}

type moon struct {
	Phase phase `json:"phase"`
}

type sunmoon struct {
	Moon moon `json:"moon"`
}

type response struct {
	Success  bool      `json:"success"`
	Response []sunmoon `json:"response"`
}

func (luna *Luna) getPhase() byte {
	response := &response{}

	resp, err := http.Get(fmt.Sprintf(url, luna.clientId, luna.clientSecret))
	if err != nil {
		log.Println("Error getting moon phase:", err)
		return 1
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading moon response: %v", err)
		return 1
	}

	err = json.Unmarshal(data, response)
	if err != nil {
		log.Printf("Error decoding moon json from \"%v\": %v", string(data), err)
		return 1
	}

	if !response.Success || len(response.Response) == 0 {
		log.Printf("Invalid moon phase response in \"%v\": %+v", string(data), response)
		return 1
	}

	// floor of x + 0.5 is round
	return byte(math.Floor(response.Response[0].Moon.Phase.Age/29*10+0.5)) % 10
}
