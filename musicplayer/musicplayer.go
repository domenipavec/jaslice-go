package musicplayer

import (
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"github.com/matematik7/jaslice-go/application"
)

type MusicPlayer struct {
	playStop chan bool
	next     chan bool
	playing  bool
	wait     chan bool
	cmd      *exec.Cmd

	playlists     []string
	playlistSongs [][]string
	playlistIndex int

	currentSong    string
	songHandler    websocket.Handler
	songWebsockets []*websocket.Conn
}

func New(config map[string]interface{}) application.Module {
	rand.Seed(time.Now().UnixNano())

	mp := &MusicPlayer{
		playStop: make(chan bool),
		wait:     make(chan bool),
		next:     make(chan bool),
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

		songs, err := filepath.Glob(filepath.Join(playlist, "*.mp3"))
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Found", len(songs), "songs in", playlist, "playlist.")

		mp.playlists = append(mp.playlists, playlist)
		mp.playlistSongs = append(mp.playlistSongs, songs)
	}

	mp.songHandler = websocket.Handler(mp.SongWebsocket)

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
		case <-mp.next:
			if mp.playing {
				mp.stopPlaying()
				mp.playSong()
			}
		}
	}
}

func (mp *MusicPlayer) playSong() {
	mp.setCurrentSong(mp.getSong())

	mp.cmd = exec.Command("omxplayer", "-o", "local", mp.currentSong)
	if err := mp.cmd.Start(); err != nil {
		log.Print(err)
	}

	go mp.waitForPlayer()
}

func (mp *MusicPlayer) stopPlaying() {
	mp.setCurrentSong("")

	if mp.cmd != nil && mp.cmd.Process != nil {
		if err := mp.cmd.Process.Kill(); err != nil {
			log.Print(err)
		}
		mp.cmd = nil
	}
}

func (mp *MusicPlayer) setCurrentSong(song string) {
	mp.currentSong = song

	for _, webSocket := range mp.songWebsockets {
		webSocket.Write([]byte(mp.currentSong))
	}
}

func (mp *MusicPlayer) waitForPlayer() {
	if err := mp.cmd.Wait(); err != nil {
		log.Print(err)
	}
	mp.wait <- true
}

func (mp *MusicPlayer) getSong() string {
	songIndex := rand.Intn(len(mp.playlistSongs[mp.playlistIndex]))
	return mp.playlistSongs[mp.playlistIndex][songIndex]
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
	} else if command == "next" {
		mp.next <- true
	} else if command == "song" {
		mp.songHandler.ServeHTTP(w, r)
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

func (mp *MusicPlayer) SongWebsocket(ws *websocket.Conn) {
	mp.songWebsockets = append(mp.songWebsockets, ws)
	ws.Write([]byte(mp.currentSong))

	for {
		time.Sleep(time.Second)
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
