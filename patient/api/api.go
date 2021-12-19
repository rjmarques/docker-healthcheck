package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/rjmarques/docker-healthcheck/patient/stress"
)

func Server(maxFix int, b *stress.CPUBurner) http.Handler {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/start", startHandler(b)).Methods("GET")
	api.HandleFunc("/stop", stoptHandler(b)).Methods("GET")
	api.HandleFunc("/healthcheck", healthcheck(maxFix))

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)

	return loggedRouter
}

func startHandler(b *stress.CPUBurner) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b.Start()
	}
}

func stoptHandler(b *stress.CPUBurner) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b.Stop()
	}
}

// naive approach: use this soft sequence to allow invokers to uniquely identify and order heackcheck requests
var sequence int64

func healthcheck(maxFix int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// simple fibonacci operation to ensure the app is not starved for CPU
		stress.Fib(maxFix)

		// output can be used as an indentifier
		w.Write([]byte(fmt.Sprintf("%d", sequence)))
		sequence++
	}
}
