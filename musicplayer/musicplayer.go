package musicplayer

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/net/websocket"

	"github.com/matematik7/jaslice-go/application"
)

// numid check with amixer controls
const volumeId = 1

type MusicPlayer struct {
	playStop chan bool
	next     chan bool
	playing  bool

	wait chan bool
	cmd  *exec.Cmd

	playlists        []string
	playlistSongs    [][]string
	playlistIndex    int
	playlistShuffles [][]int

	currentSong    string
	songHandler    websocket.Handler
	songWebsockets []*websocket.Conn

	volume int
}

func New(app *application.App, config application.Config) application.Module {
	mp := &MusicPlayer{
		playStop:  make(chan bool),
		wait:      make(chan bool),
		next:      make(chan bool),
		playlists: config.GetSliceStrings("folders"),
		volume:    100,
	}

	for _, playlist := range mp.playlists {
		songs, err := filepath.Glob(filepath.Join(playlist, "*.mp3"))
		if err != nil {
			log.Fatalln("Error globbing:", err)
		}

		log.Println("Found", len(songs), "songs in", playlist, "playlist.")

		mp.playlistSongs = append(mp.playlistSongs, songs)
		mp.playlistShuffles = append(mp.playlistShuffles, nil)
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
			}
		}
	}
}

func (mp *MusicPlayer) playSong() {
	song := mp.getSong()
	mp.setCurrentSong(song)

	mp.cmd = exec.Command("mpg321", song)

	if err := mp.cmd.Start(); err != nil {
		log.Println("Error starting:", err)
		return
	}

	go mp.waitForPlayer()
}

func (mp *MusicPlayer) stopPlaying() {
	mp.setCurrentSong("")

	if err := mp.cmd.Process.Kill(); err != nil {
		log.Println("Error killing:", err)
		return
	}
}

func (mp *MusicPlayer) setCurrentSong(song string) {
	_, mp.currentSong = filepath.Split(song)

	for _, webSocket := range mp.songWebsockets {
		webSocket.Write([]byte(mp.currentSong))
	}
}

func (mp *MusicPlayer) waitForPlayer() {
	mp.cmd.Wait()
	mp.wait <- true
}

func (mp *MusicPlayer) getSong() string {
	if len(mp.playlistShuffles[mp.playlistIndex]) == 0 {
		mp.playlistShuffles[mp.playlistIndex] = rand.Perm(len(mp.playlistSongs[mp.playlistIndex]))
	}

	songIndex := mp.playlistShuffles[mp.playlistIndex][0]
	mp.playlistShuffles[mp.playlistIndex] = mp.playlistShuffles[mp.playlistIndex][1:]

	return mp.playlistSongs[mp.playlistIndex][songIndex]
}

func (mp *MusicPlayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Path

	if command == "on" {
		mp.playStop <- true
	} else if command == "off" {
		mp.playStop <- false
	} else if command == "toggle" {
		mp.playStop <- !mp.playing
	} else if command == "next" {
		mp.next <- true
	} else if command == "song" {
		mp.songHandler.ServeHTTP(w, r)
	} else if index, ok := application.CommandInt(w, command, "playlist/", 0, len(mp.playlists)-1); ok {
		mp.playlistIndex = index
	} else if volume, ok := application.CommandInt(w, command, "volume/", 0, 100); ok {
		mp.SetVolume(volume)
	} else {
		w.WriteHeader(404)
	}
}

func (mp *MusicPlayer) SetVolume(volume int) {
	command := exec.Command("amixer", "cset", fmt.Sprintf("numid=%d", volumeId), fmt.Sprintf("%d%%", volume))
	err := command.Run()
	if err != nil {
		log.Println("Error setting volume:", err)
		return
	}
	mp.volume = volume
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

	Volume int
}

func (mp *MusicPlayer) Data() interface{} {
	return data{
		Playing: mp.playing,

		Playlists:       mp.playlists,
		CurrentPlaylist: mp.playlistIndex,

		Volume: mp.volume,
	}
}

func (mp *MusicPlayer) On() {

}

func (mp *MusicPlayer) Off() {
	mp.playStop <- false
}
