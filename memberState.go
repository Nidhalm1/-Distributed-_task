package main

type NodeState struct {
	Load   int     `json:"load"`
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Tasks  int     `json:"tasks"`
}
type message struct {
	Node  string    `json:"node"`
	State NodeState `json:"state"`
}

var state NodeState
var clusterState = make(map[string]NodeState)

type task struct {
	ID        int      `json:"id"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	Output    string   `json:"output"`
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	CreatedAt string   `json:"created_at"`
}

var mesTasks = make(map[int]task)

// rendre plutot periodique  qui tourne en arrirer pour les info
func init() {
	state.Load = 6
	state.CPU = 7
	state.Memory = 10
	state.Tasks = 10
}
