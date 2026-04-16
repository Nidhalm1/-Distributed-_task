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
			defer conn.Close()
			break
		}

		parts := strings.Fields(line)
		if len(parts) == 1 {
			fmt.Println("manque la commande")
			continue
		} // pb traiter si c des bon argmuent ou pas
		if parts[0] == "submit" {
			if len(parts) < 2 {
				fmt.Println("Usage: submit <commande> <args>")
				continue
			}
			submit := common.SubmitRequest{Type: "submit", Command: parts[1], Args: parts[2:]} // pb envoyer direct la val de commande et le serveur trait
			data, err := json.Marshal(submit)
			if err != nil {
				fmt.Println("Erreur JSON:", err)
				return
			}
			conn.Write(data)
			var r common.Response
			err = json.NewDecoder(conn).Decode(&r) // ce qu'il m'envoie
			if err != nil {
				fmt.Println("decode error:", err)
				return
			}
			fmt.Println("ID recu", r.ID)
		} else if parts[0] == "result" { // pb traiter si c des bon argmuent ou pas si c'est dans ma table
			if len(parts) < 2 {
				fmt.Println("Usage: result <id>")
				continue
			}
			result := common.SubmitRequest{Type: "result", Command: parts[1]} // pb envoyer direct la val de commande et le serveur trait
			data, err := json.Marshal(result)
			if err != nil {
				fmt.Println("Erreur JSON:", err)
				return
			}
			conn.Write(data)
			var r common.TaskResult
			err = json.NewDecoder(conn).Decode(&r) // ce qu'il m'envoie
			if err != nil {
				fmt.Println("decode error:", err)
				return
			}
			fmt.Println("ID recu", r.Status)
		}
	}
}
