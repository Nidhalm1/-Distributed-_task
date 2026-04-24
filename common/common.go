package common

import (
	"encoding/json"
	"time"
)

type Task struct {
	ID           string    `json:"id"`
	Command      string    `json:"command"`
	Args         []string  `json:"args"`
	Output       string    `json:"output"`
	Status       string    `json:"status"`
	Error        string    `json:"error"`
	CreatedAt    time.Time `json:"created_at"`
	Estimatedmem int       `json:"estimatedmem"`
	Estimatedcpu int       `json:"estimatedcpu"`
}

var tasks = make(map[string]*Task)

type Result struct {
	ID string `json:"id"`
}
type Response struct {
	ID string `json:"id"`
}

type SubmitRequest struct {
	Type         string
	ID           string
	Command      string
	Args         []string
	EstimatedCPU int
	EstimatedMem int
}
type Probe struct {
	Estimatedmem int `json:"estimatedmem"`
	Estimatedcpu int `json:"estimatedcpu"`
}
type ProbeResponse struct {
	Accepted bool
}

type TaskResult struct {
	Output string `json:"output"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
