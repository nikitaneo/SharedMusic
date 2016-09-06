package main

import (
    "fmt"
    "net/http"
    "net/url"
    "strconv"
    "os"
    "strings"
)

var edison_machines map[int64]string

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}	
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm();
	check(err)

	id, err := strconv.ParseInt(r.PostForm["id"][0], 10, 64)
	check(err)

	switch r.PostForm["command"][0] {
	case "pause":
		_, err := http.PostForm(edison_machines[id], url.Values{ "command": {"pause"} })
		check(err)
	case "unpause":
		_, err := http.PostForm(edison_machines[id], url.Values{ "command": {"unpause"} })
		check(err)
	case "add_audio":
		_, err := http.PostForm(edison_machines[id], url.Values{ "command": {"add_audio"}, "url": {r.PostForm["url"][0]} })	
		check(err)	
	}
}

func edisonHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	check(err)

	id, err := strconv.ParseInt(r.PostForm["id"][0], 10, 64)
	check(err)

	switch r.PostForm["status"][0] {
	case "connect":
		edison_machines[id] = "http://" + strings.Split(r.RemoteAddr, ":")[0] + ":8081"
		fmt.Println("connect", id)
	case "break":	
		delete(edison_machines, id)
		fmt.Println("delete", id)
	}
}

func main() {
	edison_machines = make(map[int64]string)
    http.HandleFunc("/client", clientHandler)
    http.HandleFunc("/edison", edisonHandler)
    http.ListenAndServe(":8080", nil)
}