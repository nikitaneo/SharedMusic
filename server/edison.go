package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"io/ioutil"
	"syscall"
	"io"
)

const (
	playlist_path = "/home/nikita/dev/go/src/github.com/user/sharedMusic/samples/playlist.txt"
	server_url = "http://localhost:8080/edison"
	edison_port = ":8081"
)

var mplayer *exec.Cmd

var inPipe io.WriteCloser
var outPipe io.ReadCloser
var errPipe io.ReadCloser

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm(); 
	check(err);

	command := r.PostForm["command"][0]
	switch command {
	case "pause", "unpause":
		_, err = inPipe.Write([]byte("pause\n"))
		check(err)
	case "add_audio":
		err := ioutil.WriteFile(playlist_path, []byte(r.PostForm["url"][0]), 0644)
		check(err);
	}
}

func handleCtrlC(c chan os.Signal) {
	sig := <-c
	http.PostForm(server_url, url.Values{"id": {os.Getenv("MACHINE_ID")}, "status": {"break"}})
	fmt.Println("\nsignal: ", sig)
	os.Exit(0)
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go handleCtrlC(c)

	_, err := http.PostForm(server_url, url.Values{"id": {os.Getenv("MACHINE_ID")}, "status": {"connect"}})
	check(err)

	mplayer = exec.Command("mplayer", "-slave", "-quiet", "-playlist", playlist_path)
	
	inPipe, err = mplayer.StdinPipe()
	check(err)
	outPipe, err = mplayer.StdoutPipe()
	check(err)
	errPipe, err = mplayer.StderrPipe()
	check(err)

	err = mplayer.Start()
	check(err)

	http.HandleFunc("/", handler)
	http.ListenAndServe(edison_port, nil)
}