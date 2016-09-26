// bla bla bla
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const (
	server_url = "http://localhost:8080/client"
)

func main() {
	_, err := http.PostForm(server_url, url.Values{"id": {"24864077"}, "command": {"pause"}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
