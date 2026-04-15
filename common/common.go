package common

import "time"

type Task struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Output    string    `json:"output"`
	Status    string    `json:"status"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
}

type Result struct {
	ID string `json:"id"`
}
type Request struct {
	Type string `json:"type"`
}

type TaskResult struct {
	Output string `json:"output"`
	Status string `json:"status"`
	Error  string `json:"error"`
}
