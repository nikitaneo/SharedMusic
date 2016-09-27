// Do not forget to set MACHINE_ID environment variable before compiling!

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	playlistPath = "samples/playlist.txt"
	serverHost   = "localhost:8080"
	serverPath   = "/edison"
)

type VKApiResponce struct {
	Aid       int    `json: "aid`
	Owner_id  int    `json: "owner_id"`
	Artist    string `json: "artist"`
	Title     string `json: "title"`
	Duration  int    `json: "duration"`
	Url       string `json: "url"`
	Lyrics_id string `json: "lyrics_id"`
	Genre     int    `json: "genre"`
}

type VKApiResponcesList struct {
	Responces []VKApiResponce `json:"response"`
}

type RequestToServer struct {
	Device_id int // unique id of Intel Edison
}

type ServerResponce struct {
	Command      string // pause, unpause, push, pop
	AudioID      string // [optionary] id of audio to push
	AccessTocken string // [optionary] tocken for VK API requests
}

var inPipe io.WriteCloser
var outPipe io.ReadCloser
var errPipe io.ReadCloser

var upgrader = websocket.Upgrader{}
var conn *websocket.Conn // connection to server

var mplayer *exec.Cmd

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handleCtrlC(c chan os.Signal) {
	sig := <-c
	err := conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	check(err)
	fmt.Println("\nsignal: ", sig)
	os.Exit(0)
}

func main() {
	var err error

	// mplayer always play first song in playlist.
	// we should pause mplayer at first song and on "push" command play next song
	skipAudio := true

	// create handler for Ctrl-C signal
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go handleCtrlC(c)

	// init handshake to server
	url := url.URL{Scheme: "ws", Host: serverHost, Path: serverPath}
	conn, _, err = websocket.DefaultDialer.Dial(url.String(), nil)
	check(err)
	//defer conn.Close()

	// send initial information about Intel Edison device
	id, _ := strconv.Atoi(os.Getenv("MACHINE_ID"))
	outRequest := RequestToServer{Device_id: id}
	outMessage, err := json.Marshal(outRequest)
	err = conn.WriteMessage(websocket.TextMessage, outMessage)
	check(err)

	// run mplayer in slave mode with "playlist_path" playlist
	mplayer = exec.Command("mplayer", "-slave", "-quiet", "-playlist", playlistPath)

	// handle pipes from mplayer process
	inPipe, err = mplayer.StdinPipe()
	check(err)
	outPipe, err = mplayer.StdoutPipe()
	check(err)
	errPipe, err = mplayer.StderrPipe()
	check(err)

	// start mplayer daemon
	err = mplayer.Start()
	check(err)

	// pause mplayer at first song
	_, err = inPipe.Write([]byte("pause\n"))
	check(err)

	// read messages from websocket, connected to server
	for {
		_, rawInMessage, err := conn.ReadMessage()
		check(err)

		var inMessage ServerResponce
		err = json.Unmarshal(rawInMessage, &inMessage)

		switch inMessage.Command {
		case "pause", "unpause":
			fmt.Println("pause/unpause command")

			if skipAudio {
				skipAudio = false
				_, err = inPipe.Write([]byte("pt_step 1\n"))
			} else {
				_, err = inPipe.Write([]byte("pause\n"))
				check(err)
			}
		case "push":
			fmt.Println("push command")

			apiHttpResponce, err := http.Get(
				"https://api.vk.com/method/audio.getById?audios=" +
					os.Getenv("MACHINE_ID") + "_" + inMessage.AudioID +
					"&access_token=" + inMessage.AccessTocken)
			check(err)
			apiJsonResponce, err := ioutil.ReadAll(apiHttpResponce.Body)
			check(err)

			var apiResponces = new(VKApiResponcesList)
			err = json.Unmarshal(apiJsonResponce, apiResponces)
			check(err)

			fmt.Println(apiResponces.Responces[0].Url)

			_, err = inPipe.Write([]byte("loadfile " +
				apiResponces.Responces[0].Url + " 1\n"))
			check(err)
		case "next":
			fmt.Println("next command")

			// exit from mplayer for playing with another playlist
			_, err = inPipe.Write([]byte("exit\n"))
			check(err)

			_, err = inPipe.Write([]byte("pt_step 1\n"))
			check(err)
		}

	}
}
