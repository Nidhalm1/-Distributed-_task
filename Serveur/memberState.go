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
var execQueue = make(chan common.Task, 100) //thread safe deja

var tasks = make(map[string]common.TaskResult)

var reservedCPU = 0
var reservedMEM = 0

var taskReserved = make(map[string]bool)

func init() {
	state.Load = rand.Intn(10) + 1
	state.CPU = rand.Int() * 100.0
	state.Memory = rand.Int() * 32.0
	state.Tasks = rand.Intn(20) + 1
}
