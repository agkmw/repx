package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/agkmw/workout-service/internal/app"
	"github.com/agkmw/workout-service/internal/routes"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "the port to listen to requests")
	flag.Parse()

	app, err := app.New()
	if err != nil {
		panic(err)
	}
	defer app.DB.Close()

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      routes.SetupRoutes(app),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.Logger.Info("server running", "port", port)

	if err := server.ListenAndServe(); err != nil {
		app.Logger.Error("server failed", err)
	}
}
