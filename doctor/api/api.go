package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rjmarques/docker-healthcheck/doctor/docker"
	"github.com/rjmarques/docker-healthcheck/doctor/model"
	"github.com/rjmarques/docker-healthcheck/doctor/patient"
)

type Server struct {
	mu     sync.Mutex
	checks map[string]*model.Healthcheck

	patient    *patient.Client
	dk         *docker.Client
	targetName string

	patientBurning bool
	healthErr      error
}

func NewServer(patient *patient.Client, dk *docker.Client, targetName string) *Server {
	return &Server{
		checks:     map[string]*model.Healthcheck{},
		patient:    patient,
		dk:         dk,
		targetName: targetName,
	}
}

func (s *Server) StartCollecting() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			checks, err := s.dk.GetHealthChecks()
			if err != nil {
				fmt.Printf("failed to fetch healthchecks: %v\n", err)
				s.healthErr = err
				continue
			} else {
				s.healthErr = nil
			}

			s.mu.Lock()
			for _, c := range checks {
				if s.checks[c.ID] == nil {
					fmt.Printf("new health check: %v\n", c)
					s.checks[c.ID] = c
				}
			}
			s.mu.Unlock()
		}
	}()
}

func (s *Server) Router() http.Handler {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/start", s.startHandler)
	api.HandleFunc("/stop", s.stoptHandler)
	api.HandleFunc("/metrics", s.metrics)

	r.PathPrefix("/").Handler(nocache(http.FileServer(http.Dir("static")))) // for assets in static/

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)

	return loggedRouter
}

func (s *Server) startHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.patient.Start()
	if err != nil {
		setError(w, err, http.StatusInternalServerError)
		return
	}

	s.patientBurning = true
}

func (s *Server) stoptHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.patient.Stop()
	if err != nil {
		setError(w, err, http.StatusInternalServerError)
		return
	}

	s.patientBurning = false
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	m := s.calcMetrics()
	if err := json.NewEncoder(w).Encode(m); err != nil {
		setError(w, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) calcMetrics() *model.Metrics {
	checks := s.sortedChecks()
	if len(checks) == 0 {
		return nil
	}

	last := checks[0]
	var status string
	if s.healthErr == nil {
		status = "OK"
	} else {
		status = "Error"
	}

	mean := mean(checks)
	min := min(checks)

	m := &model.Metrics{
		Status:         status,
		PatientBurning: s.patientBurning,
		LastTimming:    timming(last),
		MeanTimming:    mean,
		Prognosis:      prognosis(min, mean),
	}
	return m
}

// gets all metrics in descending order
func (s *Server) sortedChecks() []*model.Healthcheck {
	var checks []*model.Healthcheck

	s.mu.Lock()
	for _, c := range s.checks {
		checks = append(checks, c)
	}
	s.mu.Unlock()

	sort.Slice(checks, func(a, b int) bool {
		c1 := checks[a]
		c2 := checks[b]

		id1, _ := strconv.ParseInt(c1.ID, 10, 64)
		id2, _ := strconv.ParseInt(c2.ID, 10, 64)

		return id1 > id2
	})

	return checks
}

func mean(checks []*model.Healthcheck) time.Duration {
	var sumDuration time.Duration
	max := minInt(len(checks), 10) // 10 last checks is enough
	for i := 0; i < max; i++ {
		sumDuration += timming(checks[i])
	}
	return sumDuration / time.Duration(max)
}

func min(checks []*model.Healthcheck) time.Duration {
	min := timming(checks[0])
	for i := 1; i < len(checks); i++ {
		ti := timming(checks[i])
		if ti < min {
			min = ti
		}
	}
	return min
}

func prognosis(min, mean time.Duration) string {
	if float64(mean) > 1.3*float64(min) {
		return "Getting hot!"
	}
	return "All good!"
}

func timming(c *model.Healthcheck) time.Duration {
	return c.End.Sub(c.Start)
}

func setError(w http.ResponseWriter, err error, status int) {
	http.Error(w, err.Error(), status)
}

func nocache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		h.ServeHTTP(w, r)
	})
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
