// nouveau fichier execWorker.go (ou dans worker.go)
package main

import (
	"NVPROJET/common"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
)

func startExecWorkers(n int) {
	for i := 0; i < n; i++ {
		go func() {
			for task := range execQueue {
				execTask(task)
			}
		}()
	}
}
func execTask(task common.Task) {
	cmd := exec.Command(task.Command, task.Args...)
	stateMu.Lock()
	reservedCPU -= task.Estimatedcpu
	reservedMEM -= task.Estimatedmem

	stateMu.Unlock()

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
