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
	var requestType common.SubmitRequest
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)
	for {
		err := decoder.Decode(&requestType)
		if err != nil {
			fmt.Println("client disconnected or decode error:", err)
			return
		}
		switch requestType.Type {
		case "submit":
			handleSubmit(encoder, requestType)
		case "result":
			handleResult(encoder, requestType)
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

func handleResult(Encoder *json.Encoder, requestType common.SubmitRequest) {
	var id = requestType.ID
	task, ok := tasks[id]
	// gerer  le pb task nn trouvé
	if !ok {
		fmt.Println("Task not found for ID:", id)
		return
	}
	Encoder.Encode(task)

}
