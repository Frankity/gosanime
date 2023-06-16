package main

import (
	"log"
	"net/http"
	"os"

	"xyz.frankity/gosanime/main/server"
)

func main() {
	app := server.New()
	http.HandleFunc("/", app.Router.ServeHTTP)
	port := ":3000" // + os.Getenv("PORT")
	log.Println("App running again...")
	log.Fatal(http.ListenAndServe(port, nil))
}

func check(e error) {
	if e != nil {
		log.Println(e)
		os.Exit(1)
	}
}
