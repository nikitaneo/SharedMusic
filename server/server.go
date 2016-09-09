package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"strconv"
)

// upgrader for websocket connection
var upgrader = websocket.Upgrader{}

// map contains websocket for each connected Intel Edison device
var deviceConnections map[int]*websocket.Conn

type RequestToServer struct {
	Device_id int // unique id of Intel Edison
}

type ServerResponce struct {
	Command  string
	AudioURL string
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	check(err)

	id, err := strconv.Atoi(r.PostForm["id"][0])
	check(err)

	var outMessage ServerResponce
	switch r.PostForm["command"][0] {
	case "pause":
		outMessage = ServerResponce{"pause", ""}
	case "unpause":
		outMessage = ServerResponce{"unpase", ""}
		check(err)
	case "push":
		outMessage = ServerResponce{"push", r.PostForm["audioURL"][0]}
		check(err)
	case "pop":
		outMessage = ServerResponce{"pop", ""}
	}

	message, err := json.Marshal(outMessage)
	check(err)

	err = deviceConnections[id].WriteMessage(websocket.TextMessage, message)
	check(err)
}

func edisonHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	check(err)

	fmt.Println("Connect: " + string(r.RemoteAddr))

	for {
		_, rawInMessage, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var inMessage RequestToServer
		err = json.Unmarshal(rawInMessage, &inMessage)
		check(err)

		deviceConnections[inMessage.Device_id] = conn
	}

	fmt.Println("Disconnect: " + string(r.RemoteAddr))
}

func main() {
	deviceConnections = make(map[int]*websocket.Conn)
	http.HandleFunc("/client", clientHandler)
	http.HandleFunc("/edison", edisonHandler)
	http.ListenAndServe(":8080", nil)
}
