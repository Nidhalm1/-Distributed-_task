package main

import (
	"NVPROJET/common"
	"math/rand"
)

type NodeState struct {
	Load    int `json:"load"`
	CPU     int `json:"cpu"`
	Memory  int `json:"memory"`
	Tasks   int `json:"tasks"`
	PortTcp int `json:"port"`
}

var state NodeState
var clusterState = make(map[string]NodeState)

var mapAdresse = make(map[string]string)

var taskQueue = make(chan common.Task, 100) //thread safe deja
var tasks = make(map[string]common.TaskResult)

func init() {
	state.Load = rand.Intn(10) + 1
	state.CPU = rand.Int() * 100.0
	state.Memory = rand.Int() * 32.0
	state.Tasks = rand.Intn(20) + 1
	state.PortTcp = 1024 + rand.Intn(65535-1024)
}
