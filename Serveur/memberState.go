package main

import (
	"NVPROJET/common"
	"math/rand"
)

type NodeState struct {
	Load   int `json:"load"`
	CPU    int `json:"cpu"`
	Memory int `json:"memory"`
	Tasks  int `json:"tasks"`
}
type message struct {
	Node  string    `json:"node"`
	State NodeState `json:"state"`
}

var state NodeState
var clusterState = make(map[string]NodeState)

var taskQueue = make(chan common.Task, 100) //thread safe deja
var tasks = make(map[string]*common.Task)

func init() {
	state.Load = rand.Intn(10) + 1
	state.CPU = rand.Int() * 100.0
	state.Memory = rand.Int() * 32.0
	state.Tasks = rand.Intn(20) + 1
}
