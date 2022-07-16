package main

import (
	"log"
	"net/http"
	"os"

	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/server"
)

func main() {
	app := server.New()
	http.HandleFunc("/", app.Router.ServeHTTP)
	port := ":" + config.Config().Port
	log.Println("App running again...")
	log.Fatal(http.ListenAndServe(port, nil))
}

func check(e error) {
	if e != nil {
		log.Println(e)
		os.Exit(1)
	}
}
