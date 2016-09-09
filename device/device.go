package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	playlistPath = "/home/nikita/dev/go/src/github.com/user/SharedMusic/device/samples/playlist.txt"
	// server_url = "http://52.174.6.203:8080/edison"
	serverHost = "localhost:8080"
	serverPath = "/edison"
	edisonPort = ":8081"
)

type RequestToServer struct {
	Device_id int // unique id of Intel Edison
}

type ServerResponce struct {
	Command  string // pause, unpause, push, pop
	AudioURL string // [optionary] url to audio to push
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
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	check(err)
	fmt.Println("\nsignal: ", sig)
	os.Exit(0)
}

func main() {
	var err error

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
	id, _ := strconv.Atoi("24864077")
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
	if err != nil {
		fmt.Fprint(os.Stderr, "mplayer.Start() error")
	}

	// read messages from websocket, connected to server
	for {
		_, rawInMessage, err := conn.ReadMessage()
		check(err)

		var inMessage ServerResponce
		err = json.Unmarshal(rawInMessage, &inMessage)

		switch inMessage.Command {
		case "pause", "unpause":
			_, err = inPipe.Write([]byte("pause\n"))
			check(err)
		case "push":
			// ignore
		case "pop":
			// ignore
		}

	}
}
