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

type VersionedTasks struct { //permet d'identifier la verison la plus recente de l'information pour la migration (TCP est FIFO et broadcast lock donc pas besoin de plus)
	Tasks   map[string]common.Task `json:"task"`
	Version int                    `json:"version"`
}

var (
	version_actuel      = 0
	versionMu           sync.Mutex
	others_tasks_list   = make(map[string]VersionedTasks)
	others_tasks_listMu sync.RWMutex
)

type MessageSuppTask struct {
	ID_envoyeur       string `json:"id_envoyeur"`
	ID_task_concernee string `json:"task_concernee"`
	Version           int    `json:"version_data"`
}

type MessageAddTask struct {
	ID_envoyeur string      `json:"id_envoyeur"`
	Task_data   common.Task `json:"task_data"`
	Version     int         `json:"version_data"`
}

// les differents types de messages:
const (
	TypeAddTask  = "AddTask"
	TypeSuppTask = "SuppTask"
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

	case TypeAddTask:
		var msg MessageAddTask
		if err := json.Unmarshal(env.Data, &msg); err != nil {
			log.Println("Erreur décodage handleConn AddTask:", err)
			return
		}
		handleAddTask(msg)

	case TypeSuppTask:
		var msg MessageSuppTask
		if err := json.Unmarshal(env.Data, &msg); err != nil {
			log.Println("Erreur décodage handleConn SuppTask:", err)
			return
		}
		handleSuppTask(msg)

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
func handleAddTask(msg MessageAddTask) {

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

	//Ajout de la task
	tmp.Tasks[msg.Task_data.ID] = msg.Task_data
	// Mise à jour version
	tmp.Version = msg.Version

	others_tasks_list[msg.ID_envoyeur] = tmp

	fmt.Println("Je viends de ADD la task de : ", msg.ID_envoyeur)
}

func handleSuppTask(msg MessageSuppTask) {
	others_tasks_listMu.Lock()
	defer others_tasks_listMu.Unlock()

	if _, exists := others_tasks_list[msg.ID_envoyeur]; !exists {
		log.Println("node non listée handleSuppTask :", msg.ID_envoyeur)
	} else {

		tmp := others_tasks_list[msg.ID_envoyeur]
		tmp.Version = msg.Version

		if _, exists := others_tasks_list[msg.ID_envoyeur].Tasks[msg.ID_task_concernee]; !exists {
			log.Println("task non listée handleSuppTask:", msg.ID_task_concernee)
		} else {
			delete(tmp.Tasks, msg.ID_task_concernee)
		}

		others_tasks_list[msg.ID_envoyeur] = tmp

	}
}

// send:
func broadcastNode(msg common.Envelope) {
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

func broadcast_suppTask(task common.Task) {
	versionMu.Lock()
	defer versionMu.Unlock() //on lock jsuqu'ici car on envoie des mise à jour à la suite avec un numero de version

	msg := MessageSuppTask{
		ID_envoyeur:       config.Name,
		ID_task_concernee: task.ID,
		Version:           version_actuel,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Erreur marshal msg:", err)
		return
	}

	env := common.Envelope{
		Type: TypeSuppTask,
		Data: data,
	}

	broadcastNode(env)

	version_actuel++

}

func broadcast_addTask(task common.Task) {
	versionMu.Lock()
	defer versionMu.Unlock() //on lock jsuqu'à la fin car on envoie des mise à jour à la suite avec un numero de version

	msg := MessageAddTask{
		ID_envoyeur: config.Name,
		Task_data:   task,
		Version:     version_actuel,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Erreur marshal msg:", err)
		return
	}

	env := common.Envelope{
		Type: TypeAddTask,
		Data: data,
	}

	broadcastNode(env)

	version_actuel++
}
