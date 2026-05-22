//leader election : puis le partage de la liste des taches de la personne

// code serveur qui traite la liste des task des autres serveurs (pour la migration)
package main

import (
	"NVPROJET/common"
	"encoding/json"
	"fmt"
	"log"
)

type MessageTaskList struct {
	Tasks   []common.Task `json:"tasks"`
	Version int           `json:"version"`
}

const TypeTaskList = "TaskList"

// send les task: (à executer par le serveur élu)
func handleSendTask(msg_recu TaskListDemandeur, encoder *json.Encoder) {

	others_tasks_listMu.RLock()
	ID_node_remplacee := msg_recu.ID_demandeur
	fmt.Println("handleSendTask: traite celle de : ", ID_node_remplacee)

	entry, exists := others_tasks_list[ID_node_remplacee]
	if !exists { //ne peut pas arriver normalement
		log.Println("ERREUR: node inconnu (handleSendTask):", ID_node_remplacee)
		return
	}

	fmt.Println("NOMBRE DE TASK QUE JE FAIS ENVOYER: ", len(entry.Tasks))
	// Construire la liste des tasks
	taskList := make([]common.Task, 0, len(entry.Tasks))
	for _, t := range entry.Tasks {
		taskList = append(taskList, t)
	}
	version := entry.Version
	others_tasks_listMu.RUnlock()

	// Construire le message avec le count
	msg := MessageTaskList{
		Tasks:   taskList,
		Version: version,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Erreur marshal:", err)
		return
	}

	env := common.Envelope{
		Type: TypeTaskList,
		Data: data,
	}

	if err := encoder.Encode(env); err != nil {
		fmt.Println("Erreur envoi task list:", err)
	}
}

// recv les task: (pour le serveur remplacent)
func handleRecvTasks(msg MessageTaskList) {

	log.Printf("Réception de %d task dans le cadre de la migration.\n", len(msg.Tasks))

	for _, t := range msg.Tasks {

		taskQueueMapMu.Lock()
		taskQueueMap[t.ID] = t
		taskQueueMapMu.Unlock()

		taskQueue <- t
	}

}
