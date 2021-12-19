package model

import (
	"time"
)

type Status string

func (s Status) Good() bool {
	switch s {
	case "starting", "healthy":
		return true
	default:
		return false
	}
}

type Healthcheck struct {
	ID       string
	Start    time.Time
	End      time.Time
	ExitCode int
}

type Metrics struct {
	Status         string        `json:"status"`
	PatientBurning bool          `json:"patientBurning"`
	MeanTimming    time.Duration `json:"meanTimming"`
	LastTimming    time.Duration `json:"lastTimming"`
	Prognosis      string        `json:"prognosis"`
}
