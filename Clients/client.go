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
		t := common.Task{
			Command: parts[0],
			Args:    parts[1:],
		}
		data, err := json.Marshal(t)
		if err != nil {
			fmt.Println("Erreur JSON:", err)
			continue
		}
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("Erreur d'envoi:", err)
			break
		}
	}
}
