package luna

import (
	"encoding/json"
	"fmt"
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

	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(response); err != nil {
		log.Println("Error decoding moon phase:", err)
		return 1
	}

	if !response.Success || len(response.Response) == 0 {
		log.Println("Invalid response:", response)
		return 1
	}

	// floor of x + 0.5 is round
	return byte(math.Floor(response.Response[0].Moon.Phase.Age/29*10+0.5)) % 10
}
