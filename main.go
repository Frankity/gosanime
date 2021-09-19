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

	log.Println("App running..")
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))

}

func check(e error) {
	if e != nil {
		log.Println(e)
		os.Exit(1)
	}
}
