package main

import (
	"fmt"
	"net/http"
	"os"
)

const (
	host = "localhost"
	port = ":8899"
)

func main() {
	fmt.Fprintf(os.Stdout, "\nTile server was runnig on %s%s", host, port)

	http.HandleFunc("/tiles/", TileHandler)
	http.HandleFunc("/height/", DemHandler)
	http.HandleFunc("/profile", ProfileHandler)

	http.ListenAndServe(port, nil)
}
