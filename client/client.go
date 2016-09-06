package main

import (
	"net/http"
	"net/url"
	"os"
	"fmt"
)

func main() {
	_, err := http.PostForm("http://localhost:8080/client", url.Values{"id": {"24864077"}, "command": {"pause"}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}	
}