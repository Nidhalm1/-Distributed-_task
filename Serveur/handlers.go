package main

import (
	"NVPROJET/common"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

var stateMu sync.Mutex

// pb veriefer l'utiliser de list
func handleClient(conn net.Conn) {
	defer conn.Close() // ← ça règle tout
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)
	var env common.Envelope
	for {
		err := decoder.Decode(&env)
		if err != nil {
			fmt.Println("client disconnected", err)
			return
		}
		switch env.Type {
		case "submit": //client
			var req common.SubmitRequest
			json.Unmarshal(env.Data, &req)
			handleSubmit(encoder, req)
		case "result": //client
			var req common.Result
			json.Unmarshal(env.Data, &req)
			handleResult(encoder, req)
		case "TaskResult": //serveur
			var taskresult common.TaskResult
			json.Unmarshal(env.Data, &taskresult)
			handleTaskResult(taskresult)

		case "Task": //serveur
			var task common.Task
			json.Unmarshal(env.Data, &task)
			go handleTask(task)
		case "Probe": //serveur
			stateMu.Lock()
			var probe common.Probe
			json.Unmarshal(env.Data, &probe)
			var accept = handleProbe(probe)
			encoder.Encode(common.ProbeResponse{Accepted: accept})
			stateMu.Unlock()
		default:
		}
	}
}

func handleSubmit(encoder *json.Encoder, requestType common.SubmitRequest) {
	var t common.Task
	t.Estimatedcpu = requestType.EstimatedCPU
	t.Estimatedmem = requestType.EstimatedMem
	t.Command = requestType.Command
	t.Args = requestType.Args
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now().UTC()
	t.Status = "wait"
	tasks[t.ID] = common.TaskResult{}
	var resp common.Response = common.Response{
		ID: t.ID,
	}
	taskQueueMapMu.Lock()
	taskQueueMap[t.ID] = t
	taskQueueMapMu.Unlock()

	taskQueue <- t

	encoder.Encode(resp)
	fmt.Println("Task reçue:", t.Command, t.Args)

	/// on va mettre à jour notre file de task chez tous les serveurs:
	log.Printf("%s: broadcast suite à l'ajout d'une task d'une task\n", config.Name)
	go broadcastMyTasks()
	///
}

func handleResult(Encoder *json.Encoder, resultRequest common.Result) {
	var id = resultRequest.ID
	task, ok := tasks[id]
	// gerer  le pb task nn trouvé
	if !ok {
		fmt.Println("Task not found for ID:", id)
		return
	}
	Encoder.Encode(task)
}
func handleTaskResult(taskresult common.TaskResult) {
	// regarde si elle m'appartient
	var id = taskresult.ID
	tasks[id] = taskresult
}

func handleTask(task common.Task) {
	taskReserved[task.ID] = true
	execQueue <- task
}

func handleProbe(probe common.Probe) bool {
	accepted := state.CPU-reservedCPU >= probe.Estimatedcpu && state.Memory-reservedMEM >= probe.Estimatedmem
	if accepted {
		reservedCPU += probe.Estimatedcpu
		reservedMEM += probe.Estimatedmem
		go func() {
			time.Sleep(2 * time.Second) // si je recois pas apres 2s
			if !taskReserved[probe.ID] {
				reservedCPU -= probe.Estimatedcpu
				reservedMEM -= probe.Estimatedmem
			}
		}()
	}
	return accepted
}
