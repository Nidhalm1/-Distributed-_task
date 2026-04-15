package main

import (
	"NVPROJET/common"
)

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

var taskQueue = make(chan common.Task, 100) //thread safe deja
var tasks = make(map[string]common.Task)

// rendre plutot periodique  qui tourne en arrirer pour les info
func init() {
	state.Load = 6
	state.CPU = 7
	state.Memory = 10
	state.Tasks = 10
}
