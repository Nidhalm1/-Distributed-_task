package main

import (
	"NVPROJET/common"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var tasks = make(map[string]*common.TaskResult)

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
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Entrez une commande (ou 'exit' pour quitter) :")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if strings.TrimSpace(line) == "exit" {
			defer conn.Close()
			break
		}

		parts := strings.Fields(line)
		if parts[0] == "submit" {
			if len(parts) < 4 {
				fmt.Println("Usage: submit cpu=<nb> mem=<nb> <commande> <args>")
				continue
			}
			cpuStr := strings.TrimPrefix(parts[1], "cpu=")
			cpu, err := strconv.Atoi(cpuStr)
			memStr := strings.TrimPrefix(parts[2], "mem=")
			mem, err := strconv.Atoi(memStr)
			submit := common.SubmitRequest{Type: "submit", EstimatedCPU: cpu, EstimatedMem: mem, Command: parts[3], Args: parts[4:]}
			encoder.Encode(submit)
			var r common.Response
			err = decoder.Decode(&r)
			if err != nil {
				fmt.Println("decode error:", err)
				return
			}
			tasks[r.ID] = &common.TaskResult{}
			fmt.Println("ID recu", r.ID)
		} else if parts[0] == "result" { // pb traiter si c des bon argmuent ou pas si c'est dans ma table
			// faudra tester si elle a deja ete mis a jou sans interroger direcment le serveur
			if len(parts) < 2 {
				fmt.Println("Usage: result <id>")
				continue
			}
			if _, ok := tasks[parts[1]]; !ok { /*si id existe*/
				fmt.Println("ID inconnu :", parts[1])
				continue
			}
			result := common.SubmitRequest{Type: "result", ID: parts[1]} // pb envoyer direct la val de commande et le serveur trait
			encoder.Encode(result)
			var r common.TaskResult
			err = decoder.Decode(&r) // ce qu'il m'envoie
			if err != nil {
				fmt.Println("decode error:", err)
				return
			}
			tasks[result.ID] = &r
			fmt.Println("ID recu", r.Status)
		}
	}
}
