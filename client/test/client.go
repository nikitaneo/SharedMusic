package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const (
	serverHost = "http://localhost:8080"
	serverPath = "/client"
)

var c = flag.String("c", "pause", "Command to send to device: pause/push/next")

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	switch *c {
	case "pause":
		_, err := http.PostForm(serverHost+serverPath, url.Values{"id": {"24864077"}, "command": {"pause"}})
		check(err)
	case "push":
		_, err := http.PostForm(serverHost+serverPath, url.Values{"id": {"24864077"}, "command": {"push"}, "AudioID": {"456239046"}})
		check(err)
	case "next":
		_, err := http.PostForm(serverHost+serverPath, url.Values{"id": {"24864077"}, "command": {"next"}})
		check(err)
	}
}
