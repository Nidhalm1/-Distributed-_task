// code serveur qui traite la liste des task des autres serveurs (pour la migration)
package main

import (
	"NVPROJET/common"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
)

type VersionedTasks struct { //permet d'identifier la verison la plus recente de l'information pour la migration (TCP est FIFO)
	Tasks   map[string]common.Task `json:"task"`
	Version int                    `json:"version"`
}

var (
	version_actuel      = 0
	versionMu           sync.Mutex
	others_tasks_list   = make(map[string]VersionedTasks)
	others_tasks_listMu sync.RWMutex

	boradcastMutex sync.RWMutex //Même si cela baisse les performances, il vaut mieux preserver un "version_actuel" coherent en utilisant fifo plutôt que sa valeur
)

type MessageMajTask struct {
	ID_envoyeur string                 `json:"id_envoyeur"`
	Tasks       map[string]common.Task `json:"task_data"`
	Version     int                    `json:"version_data"`
}

// les differents types de messages:
const (
	TypeMajTask = "MajTask"
)

func startServerTask(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("startServerTask : ", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnServeurTask(conn)
	}
}

func handleConnServeurTask(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)

	var env common.Envelope
	if err := decoder.Decode(&env); err != nil {
		log.Println("Erreur décodage envelope:", err)
		return
	}

	switch env.Type {

	case TypeMajTask:
		var msg MessageMajTask
		if err := json.Unmarshal(env.Data, &msg); err != nil {
			log.Println("Erreur décodage handleConn AddTask:", err)
			return
		}
		handleMajdTask(msg)

	case TypeElection:
		var msg ElectionMessage
		if err := json.Unmarshal(env.Data, &msg); err != nil {
			log.Println("Erreur décodage handleConn SuppTask:", err)
			return
		}
		recv_election(msg)

	case TypeRepElection:
		var msg ElectionMessage
		if err := json.Unmarshal(env.Data, &msg); err != nil {
			log.Println("Erreur décodage handleConn SuppTask:", err)
			return
		}
		handleElection(msg)
	case TypeAskTasks:
		var msg TaskListDemandeur
		if err := json.Unmarshal(env.Data, &msg); err != nil {
			log.Println("Erreur décodage handleConn SuppTask:", err)
			return
		}
		encoder := json.NewEncoder(conn)
		handleSendTask(msg, encoder)
	default:
		log.Println("Type inconnu:", env.Type)
	}
}

// recv :

func handleMajdTask(msg MessageMajTask) {
	others_tasks_listMu.Lock()
	defer others_tasks_listMu.Unlock()

	// si nouvelle envoyeur
	if _, exists := others_tasks_list[msg.ID_envoyeur]; !exists {

		others_tasks_list[msg.ID_envoyeur] = VersionedTasks{
			Tasks:   make(map[string]common.Task),
			Version: msg.Version,
		}
	}

	tmp := others_tasks_list[msg.ID_envoyeur]

	//on est obligé de faire une copie profonde:
	newTasks := make(map[string]common.Task, len(msg.Tasks))
	for k, v := range msg.Tasks {
		newTasks[k] = v
	}
	tmp.Tasks = newTasks
	tmp.Version = msg.Version

	others_tasks_list[msg.ID_envoyeur] = tmp

	log.Printf("%s: Je viends de mettre à jour les task de : %s ", config.Name, msg.ID_envoyeur)
}

// send:
func broadcastToNodes(msg common.Envelope) {
	boradcastMutex.Lock()
	defer boradcastMutex.Unlock()
	for nodeName, node := range clusterState {

		if nodeName == config.Name {
			continue
		}
		target := fmt.Sprintf("%s:%d", mapAdresse[nodeName], node.PortTcpDataTask)

		conn, err := net.Dial("tcp", target)
		if err != nil {
			log.Println("Erreur connexion:", err)
			continue
		}

		json.NewEncoder(conn).Encode(msg)
		conn.Close()
	}

}

func broadcastMyTasks() {
	// Vider taskQueue sans bloquer et collecter les tâches
	versionMu.Lock()

	version_actuel++

	taskQueueMapMu.Lock()

	msg := MessageMajTask{
		ID_envoyeur: config.Name,
		Tasks:       taskQueueMap,
		Version:     version_actuel,
	}
	taskQueueMapMu.Unlock()

	versionMu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Erreur marshal broadcastOwnTasks:", err)
		return
	}

	env := common.Envelope{
		Type: TypeMajTask,
		Data: data,
	}

	broadcastToNodes(env)
}
