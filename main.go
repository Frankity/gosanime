package main

import (
	"log"
	"net/http"
	"os"

	"xyz.frankity/gosanime/main/server"
)

// main is the entry point of the Gosanime application.
// It initializes the server, sets up the HTTP handler to use the server's router,
// and starts listening for HTTP requests on port 3000.
// It will log a fatal error if the server fails to start.
func main() {
	app := server.New()
	http.HandleFunc("/", app.Router.ServeHTTP)
	port := ":3000" // Port for the server to listen on.
	log.Println("App running again...")
	log.Fatal(http.ListenAndServe(port, nil))
}

// check is a helper function that logs the error and exits the program with status code 1
// if the provided error is not nil.
// This function is typically used for handling critical errors that prevent the application
// from starting or running correctly.
//
// Parameters:
//   e: The error to check. If nil, the function does nothing.
func check(e error) {
	if e != nil {
		log.Println(e)
		os.Exit(1)
	}
}
