package musicplayer

import (
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/matematik7/jaslice-go/application"
)

type MusicPlayer struct {
	playStop chan bool
	playing  bool
	wait     chan bool
	cmd      *exec.Cmd

	playlists     []string
	playlistIndex int
}

func New(config map[string]interface{}) application.Module {
	mp := &MusicPlayer{
		playStop: make(chan bool),
		wait:     make(chan bool),
	}

	folders, success := config["folders"].([]interface{})
	if !success {
		log.Fatal("Folders must be a list")
	}

	for _, folder := range folders {
		playlist, success := folder.(string)
		if !success {
			log.Fatal("Folder must be a string")
		}

		mp.playlists = append(mp.playlists, playlist)
	}

	go mp.control()

	return mp
}

func (mp *MusicPlayer) control() {
	for {
		select {
		case ps := <-mp.playStop:
			if !mp.playing && ps {
				mp.playing = true
				mp.playSong()
			} else if mp.playing && !ps {
				mp.playing = false
				mp.stopPlaying()
			}
		case <-mp.wait:
			if mp.playing {
				mp.playSong()
			}
		}
	}
}

func (mp *MusicPlayer) playSong() {
	song := mp.getSong()

	mp.cmd = exec.Command("omxplayer", "-o", "local", song)
	if err := mp.cmd.Start(); err != nil {
		log.Print(err)
	}

	go mp.waitForPlayer()
}

func (mp *MusicPlayer) stopPlaying() {
	if mp.cmd != nil && mp.cmd.Process != nil {
		if err := mp.cmd.Process.Kill(); err != nil {
			log.Print(err)
		}
		mp.cmd = nil
	}
}

func (mp *MusicPlayer) waitForPlayer() {
	if err := mp.cmd.Wait(); err != nil {
		log.Print(err)
	}
	mp.wait <- true
}

func (mp *MusicPlayer) getSong() string {
	return "testsong"
}

type status struct {
	Playing bool `json:"playing"`
}

func (mp *MusicPlayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Path

	if command == "play" {
		mp.playStop <- true
	} else if command == "stop" {
		mp.playStop <- false
	} else if strings.HasPrefix(command, "playlist/") {
		newIndex, err := strconv.Atoi(command[9:])
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		if newIndex >= len(mp.playlists) || newIndex < 0 {
			log.Println("Index out of range")
			w.WriteHeader(400)
			return
		}

		mp.playlistIndex = newIndex
	} else {
		w.WriteHeader(404)
	}
}

type data struct {
	Playing bool

	Playlists       []string
	CurrentPlaylist int
}

func (mp *MusicPlayer) Data() interface{} {
	return data{
		Playing: mp.playing,

		Playlists:       mp.playlists,
		CurrentPlaylist: mp.playlistIndex,
	}
}
