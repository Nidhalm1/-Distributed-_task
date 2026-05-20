package main

import (
	"NVPROJET/common"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
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
			fmt.Println("client disconnected or decode error:", err)
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
	taskQueue <- t
	encoder.Encode(resp)
	fmt.Println("Task reçue:", t.Command, t.Args)
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
	cmd := exec.Command(task.Command, task.Args...)

	output, err := cmd.CombinedOutput() // stdout + stderr ensemble

	result := common.TaskResult{
		ID:     task.ID,
		Output: string(output),
	}
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
	} else {
		result.Status = "done"
	}

	// renvoyer le résultat au dispatcher
	var addr string
	if net.ParseIP(task.ResultAddr) != nil && net.ParseIP(task.ResultAddr).To4() == nil {
		// IPv6
		addr = fmt.Sprintf("[%s]:%d", task.ResultAddr, task.ResultPort)
	} else {
		// IPv4 or hostname
		addr = fmt.Sprintf("%s:%d", task.ResultAddr, task.ResultPort)
	}
	fmt.Println("Envoi du résultat à l'adresse :", task.ResultPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("impossible de contacter le dispatcher")
		return
	}
	defer conn.Close()

	data, _ := json.Marshal(result)
	env := common.Envelope{Type: "TaskResult", Data: data}
	json.NewEncoder(conn).Encode(env)

}

func handleProbe(probe common.Probe) bool {
	accepted := state.CPU >= probe.Estimatedcpu && state.Memory >= probe.Estimatedmem
	if accepted {
		state.CPU -= probe.Estimatedcpu
		state.Memory -= probe.Estimatedmem
	}
	return accepted
}
