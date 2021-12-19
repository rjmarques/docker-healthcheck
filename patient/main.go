package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/rjmarques/docker-healthcheck/patient/api"
	"github.com/rjmarques/docker-healthcheck/patient/stress"
)

func main() {
	numThreads := runtime.NumCPU()
	burner := stress.NewCPUBurner(numThreads)

	maxFib := findRightFib()
	fmt.Printf("the right fib number for your CPU is %d\n", maxFib)
	sv := api.Server(maxFib, burner)
	log.Fatal(http.ListenAndServe(":80", sv))
}

// tries to find the first fib number that takes > 400ms to find
func findRightFib() int {
	var N = 2
	for {
		start := time.Now()
		stress.Fib(N)
		end := time.Now()

		dur := end.Sub(start)
		if dur >= 400*time.Millisecond {
			return N
		}
		N++
	}
}
