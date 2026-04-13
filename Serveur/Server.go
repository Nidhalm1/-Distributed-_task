package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/memberlist"
)

var idt int = 0

// factoriser en plusieur fonctions
func boucle(list *memberlist.Memberlist) {
	scanner := bufio.NewScanner((os.Stdin))

	// go et implementer un serveuur qui ecoute en c++ et grpc ou autre
	for {
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()
		result := strings.Fields(input)
		if len(result) == 0 {
			continue
		}
		if result[0] == "submit" {
			task := Task{ID: idt, Command: result[1], Args: result[2:]}
			idt++

			var mini = 100.0
			var node = ""
			for key, val := range clusterState {
				if val.CPU < mini {
					mini = val.CPU
					node = key
				}
			}
			nodes := list.Members() // inclue moi
			// verfier que c'est pas moi je m'envoie pas à moi mm
			//car dans cluster les noeud peuvnet m'envoye ma valeur
			var targetNode *memberlist.Node
			for _, n := range nodes {
				if n.Name == node {
					targetNode = n
					break
				}
			}
			data, _ := json.Marshal(task)
			if targetNode != nil && targetNode.Name != config.Name {
				if err := list.SendReliable(targetNode, data); err != nil {
					fmt.Printf("Error sending task to node %s: %v\n", node, err)
				}
			} else if targetNode != nil && targetNode.Name == config.Name {
				config.Delegate.NotifyMsg(data) //regler mieux le code car notifiy p etre bcp de chose genre if
			} else {
				fmt.Println("No suitable node found.")
			}
		}
	}
}
