package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rjmarques/docker-healthcheck/doctor/api"
	"github.com/rjmarques/docker-healthcheck/doctor/docker"
	"github.com/rjmarques/docker-healthcheck/doctor/patient"
)

func main() {
	targetName := "patient"

	pc := patient.NewClient(targetName)
	dc, err := docker.NewClient("localhost", targetName)
	if err != nil {
		panic(err)
	}
	fmt.Println("connected to Docker!")

	sv := api.NewServer(pc, dc, targetName)
	sv.StartCollecting()
	log.Fatal(http.ListenAndServe(":8080", sv.Router()))
}
