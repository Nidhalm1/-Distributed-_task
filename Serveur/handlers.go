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

func handleSubmit(conn net.Conn, list *memberlist.Memberlist) {
	defer conn.Close()
	var t common.Task
	err := json.NewDecoder(conn).Decode(&t) // ce qu'il m'envoie
	if err != nil {
		fmt.Println("decode error:", err)
		return
	}
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now().UTC()
	taskQueue <- t
	resultID, err := json.Marshal(t.ID)
	if err != nil {
		fmt.Println("Erreur JSON:", err)
	}
	_, err = conn.Write(resultID)
	if err != nil {
		fmt.Println("Erreur d'envoi:", err)
	}
	fmt.Println("Task reçue:", t.Command, t.Args)
}

func handleClient(conn net.Conn, list *memberlist.Memberlist) {
	var requestType common.Request
	for {
		err := json.NewDecoder(conn).Decode(&requestType) // ce qu'il m'envoie
		if err != nil {
			fmt.Println("decode error:", err)
			return
		}
		switch requestType.Type {
		case "submit":
			handleSubmit(conn, list)
		default:
			handleResult(conn, list)
		}
	}
}

func handleResult(conn net.Conn, list *memberlist.Memberlist) {
	defer conn.Close()
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
