package main

import (
	"NVPROJET/common"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

// pb veriefer l'utiliser de list
func handleClient(conn net.Conn) {
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
		case "submit":
			var req common.SubmitRequest
			json.Unmarshal(env.Data, &req)
			handleSubmit(encoder, req)
		case "result":
			var req common.Result
			json.Unmarshal(env.Data, &req)
			handleResult(encoder, req)
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
