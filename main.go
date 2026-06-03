package main

import (
	"flag"
	"net/http"

	"xyz.frankity/gosanime/main/logger"
	"xyz.frankity/gosanime/main/server"
)

func main() {
	logToFile := flag.Bool("logfile", false, "write logs to logs/gosanime-YYYY-MM-DD.log")
	flag.Parse()

	cleanup, err := logger.Init(*logToFile)
	if err != nil {
		logger.Fatal("failed to initialize logger", "err", err)
	}
	defer cleanup()

	app := server.New()
	http.HandleFunc("/", app.Router.ServeHTTP)

	port := ":3000"
	logger.L.Info("server started", "port", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		logger.Fatal("server error", "err", err)
	}
}
