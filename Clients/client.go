package main

import (
	"NVPROJET/common"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <adresse:port>")
		return
	}
	address := os.Args[1]
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Erreur de connexion:", err)
		return
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Entrez une commande (ou 'exit' pour quitter) :")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if strings.TrimSpace(line) == "exit" {
			break
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		if parts[0] == "submit" {
			requestType := common.Request{
				Type: "submit",
			}
			requestTyped, err := json.Marshal(requestType)
			if err != nil {
				fmt.Println("Erreur JSON:", err)
				continue
			}
			_, err = conn.Write(requestTyped)
			if err != nil {
				fmt.Println("Erreur d'envoi:", err)
				break
			}
			t := common.Task{
				Command: parts[1],
				Args:    parts[2:],
			}
			tdata, err := json.Marshal(t)
			if err != nil {
				fmt.Println("Erreur JSON:", err)
				continue
			}
			_, err = conn.Write(tdata)
			if err != nil {
				fmt.Println("Erreur d'envoi:", err)
				break
			}
			reader := bufio.NewReader(conn) // je recois forcement Lid apres
			line, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Println("Erreur de lectur id recu:", err)
				break
			}
			var result common.Result
			err = json.Unmarshal(line, &result)
			fmt.Println("id generer", result.ID)

		} else { // pb gerer les exit ect
			if len(parts) < 2 {
				fmt.Println("Usage: result <id>")
				continue
			}
			requestType := common.Request{
				Type: "result",
			}
			requestTyped, err := json.Marshal(requestType)
			if err != nil {
				fmt.Println("Erreur JSON:", err)
				continue
			}
			_, err = conn.Write(requestTyped)
			if err != nil {
				fmt.Println("Erreur d'envoi:", err)
				break
			}
			demande := common.Result{
				ID: parts[1],
			}
			Resultdata, err := json.Marshal(demande)
			if err != nil {
				fmt.Println("Erreur JSON:", err)
				continue
			}
			_, err = conn.Write(Resultdata)
			if err != nil {
				fmt.Println("Erreur d'envoi:", err)
				break
			}
			reader := bufio.NewReader(conn) // je recois forcement  apres
			line, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Println("Erreur de lectur id recu:", err)
				break
			}
			var task common.TaskResult
			json.Unmarshal(line, &task)
			fmt.Printf("Résultat pour la tâche %s :\nSortie: %s\nErreur: %s\nStatut: %s\n", parts[1], task.Output, task.Error, task.Status)
		}
	}
}
