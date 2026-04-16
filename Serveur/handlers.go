package main

import (
	"NVPROJET/common"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/memberlist"
)

// pb veriefer l'utiliser de list
func handleClient(conn net.Conn, list *memberlist.Memberlist) {
	var requestType common.SubmitRequest
	for {
		err := json.NewDecoder(conn).Decode(&requestType) // ce qu'il m'envoie
		if err != nil {
			fmt.Println("decode error:", err)
			return
		}
		switch requestType.Type {
		case "submit":
			handleSubmit(conn, list, requestType)
		case "result":
			handleResult(conn, list, requestType)
		default:
		}
	}
}

func handleSubmit(conn net.Conn, list *memberlist.Memberlist, requestType common.SubmitRequest) {
	var t common.Task
	t.Command = requestType.Command
	t.Args = requestType.Args
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now().UTC()
	taskQueue <- t
	tasks[t.ID] = &t
	var resp common.Response = common.Response{
		ID: t.ID,
	}
	resultID, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("Erreur JSON:", err)
	}
	_, err = conn.Write(resultID)
	if err != nil {
		fmt.Println("Erreur d'envoi:", err)
	}
	fmt.Println("Task reçue:", t.Command, t.Args)
}

func handleResult(conn net.Conn, list *memberlist.Memberlist, requestType common.SubmitRequest) {
	var r common.Result
	err := json.NewDecoder(conn).Decode(&r) // ce qu'il m'envoie
	if err != nil {
		fmt.Println("decode error:", err)
		return
	}
	task, ok := tasks[r.ID]
	// gerer  le pb task nn trouvé
	if !ok {
		fmt.Println("Task not found for ID:", r.ID)
		return
	}
	resp := common.TaskResult{
		Output: task.Output,
		Status: task.Status,
		Error:  task.Error,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("Erreur JSON:", err)
		return
	}
	conn.Write(data)
}
