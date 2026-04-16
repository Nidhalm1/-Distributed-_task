package common

import (
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Output    string    `json:"output"`
	Status    string    `json:"status"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
}

var tasks = make(map[string]*Task)

type Result struct {
	ID string `json:"id"`
}
type Response struct {
	ID string `json:"id"`
}

type SubmitRequest struct {
	Type    string
	Command string
	Args    []string
}

type TaskResult struct {
	Output string `json:"output"`
	Status string `json:"status"`
	Error  string `json:"error"`
}
