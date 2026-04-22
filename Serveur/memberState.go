package main

import (
	"NVPROJET/common"
	"math/rand"
	"time"
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
var tasks = make(map[string]*common.Task)

func init() {
	rand.Seed(time.Now().UnixNano())
	state.Load = rand.Intn(10) + 1
	state.CPU = rand.Float64() * 100.0
	state.Memory = rand.Float64() * 32.0
	state.Tasks = rand.Intn(20) + 1
}
