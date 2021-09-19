package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"xyz.frankity/gosanime/main/server"
)

func main() {
	app := server.New()

	http.HandleFunc("/", app.Router.ServeHTTP)

	os.Setenv("PORT", "80")
	log.Println("App running..")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))

}

func check(e error) {
	if e != nil {
		log.Println(e)
		os.Exit(1)
	}
}
