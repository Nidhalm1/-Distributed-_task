package main

//  submit cpu=10 mem=15 ls -l

import (
	"NVPROJET/common"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var tasks = make(map[string]*common.TaskResult)

var reconnectDelay = 5 * time.Second // temps avant reconnexion si la connexion est coupée côté serveur

func connect(address string) net.Conn {
	for {
		fmt.Println("Tentative de connexion en cours ...")
		conn, err := net.Dial("tcp", address)
		if err != nil {
			fmt.Println("Tentative de connexion échouée. Nouvelle tentative dans", reconnectDelay, " secondes")
			time.Sleep(reconnectDelay)
			continue
		}
		fmt.Println("La connexion a été avec succés établie avec :", address)

		return conn
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <adresse:port>")
		return
	}
	address := os.Args[1]
	conn := connect(address)

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
			submit := common.SubmitRequest{EstimatedCPU: cpu, EstimatedMem: mem, Command: parts[3], Args: parts[4:]}
			data, _ := json.Marshal(submit)
			encoder.Encode(common.Envelope{Type: "submit", Data: data})
			var r common.Response
			err = decoder.Decode(&r)
			if err != nil {
				fmt.Println("Connexion perdue. Veuillez patienter...")
				conn = connect(address)
				encoder = json.NewEncoder(conn)
				decoder = json.NewDecoder(conn)
				continue
			}
			tasks[r.ID] = &common.TaskResult{}
			fmt.Println("ID recu", r.ID)
		} else if parts[0] == "result" { // pb traiter si c des bon argmuent ou pas si c'est dans ma table
			// faudra tester si elle a deja ete mis a jou sans interroger direcment le serveur
			if len(parts) < 2 {
				fmt.Println("Usage: result <id>")
				continue
			}
			if _, ok := tasks[parts[1]]; !ok { /*si id existe pas en continue*/
				fmt.Println("ID inconnu :", parts[1])
				continue
			}
			result := common.Result{ID: parts[1]}
			data, _ := json.Marshal(result)
			encoder.Encode(common.Envelope{Type: "result", Data: data})
			var r common.TaskResult
			err := decoder.Decode(&r) // ce qu'il m'envoie
			if err != nil {
				fmt.Println("Connexion perdue. Veuillez patienter...")
				conn = connect(address)
				encoder = json.NewEncoder(conn)
				decoder = json.NewDecoder(conn)
				continue
			}
			tasks[result.ID] = &r
			fmt.Println("etat ID recu", r.Output)
		}
	}
}
