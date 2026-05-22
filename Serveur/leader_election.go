//Calcul de maximum
// Idée:
// Diffusion de son identité et attente des identités de tous les autres
// Le processus d’identifiant max devient le chef

package main

import (
	"NVPROJET/common"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const TypeElection = "Election"
const TypeRepElection = "RepElection"

// pour la demande de task list:
const TypeAskTasks = "AskTasks"

type TaskListDemandeur struct {
	ID_demandeur string `json:"id_demandeur"`
}

///

var (
	known_ids                  = make(map[string]int)
	known_idsMu                sync.Mutex
	nbr_serveur_attente_de_rep int

	electionTimerDone = false
	electionTimerMu   sync.Mutex
)

type ElectionMessage struct {
	ID_envoyeur        string `json:"id_envoyeur"`
	Version_enregistre int    `json:"version"`
}

// pour les autres serveurs:
func recv_election(msg ElectionMessage) {
	log.Println("Election: demande de vote recu chez", config.Name, "-", mapAdresse[msg.ID_envoyeur], ":", clusterState[msg.ID_envoyeur].PortTcpDataTask)

	target := fmt.Sprintf("%s:%d", mapAdresse[msg.ID_envoyeur], clusterState[msg.ID_envoyeur].PortTcpDataTask)
	conn, err := net.Dial("tcp", target)

	if err != nil {
		log.Println("Erreur connexion election recv_election():", err)
		return
	}

	var version = 0

	if data, exists := others_tasks_list[msg.ID_envoyeur]; exists {
		version = data.Version
	}

	rep := ElectionMessage{
		ID_envoyeur:        config.Name,
		Version_enregistre: version,
	}

	data, err := json.Marshal(rep)
	if err != nil {
		log.Println("Erreur marshal election:", err)
		return
	}

	env := common.Envelope{
		Type: TypeRepElection,
		Data: data,
	}

	err = json.NewEncoder(conn).Encode(env)
	if err != nil {
		log.Println("erreur encode election dans recv_election:", err)
	}
}

//pour l'initateur:

// initiation: le serveur envoie son id à tous les atres serveurs
func broadcast_election() {

	electionTimerDone = false

	time.Sleep(3 * time.Second)                          //on se laisse le temps de se mettre en route
	nbr_serveur_attente_de_rep = (len(clusterState) - 1) //"-1" car on est present dans la liste

	msg := ElectionMessage{
		ID_envoyeur:        config.Name,
		Version_enregistre: 0, //cette valeur n'aura pas d'importances
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Erreur marshal election:", err)
		return
	}

	env := common.Envelope{
		Type: TypeElection,
		Data: data,
	}

	// Diffuser à tous les noeuds
	for nodeName, node := range clusterState {
		if nodeName == config.Name {
			continue
		}
		target := fmt.Sprintf("%s:%d", mapAdresse[nodeName], node.PortTcpDataTask)
		conn, err := net.Dial("tcp", target)
		if err != nil {
			log.Println("Erreur connexion election broadcast_election():", err)
			nbr_serveur_attente_de_rep--
			continue
		}

		json.NewEncoder(conn).Encode(env)
		conn.Close()
	}

	// Lancer le timer — si electionTimerDone n'est pas true au bout de 10 secondes, on appelle choixLeader
	go func() {
		time.Sleep(10 * time.Second)
		electionTimerMu.Lock()
		defer electionTimerMu.Unlock()
		if !electionTimerDone {
			log.Println("Election: timeout, je prends la décision avec les réponses reçues")
			choixLeader()
		}
	}()

	log.Println("Election: Fin du broadcast sur tous les serveur. J'attends : ", nbr_serveur_attente_de_rep)
}

// on recoit les reponse à notre demande:
func handleElection(msg ElectionMessage) {
	electionTimerMu.Lock()
	if electionTimerDone == true { //cette reponse a été recu trop tardivement
		return
	}

	log.Println("Election: reponse à notre demande de vote recu :", config.Name)
	known_idsMu.Lock()
	known_ids[msg.ID_envoyeur] = msg.Version_enregistre
	total := len(known_ids)
	known_idsMu.Unlock()

	log.Printf("Election: reçu rep. de %s (%d connus sur %d)\n",
		msg.ID_envoyeur, total, nbr_serveur_attente_de_rep)

	// PHASE 3 — Tous les IDs reçus → calculer le max
	if total == nbr_serveur_attente_de_rep {

		//on retire le timer:
		electionTimerDone = true
		electionTimerMu.Unlock()
		//
		choixLeader()
	} else {
		electionTimerMu.Unlock()
	}
}

func choixLeader() {
	known_idsMu.Lock()
	defer known_idsMu.Unlock()

	var leader = config.Name
	var maxVersion = 0
	for id, version := range known_ids {
		if maxVersion < version {
			leader = id
		}
	}

	//Résultat
	if leader == config.Name { //cas où aucune version n'a été envoyer aux autres:
		log.Println("Election: Je suis le chef ! Je commence imediatement mes taches.")
	} else {
		log.Printf("Election: Le chef est : %s. Je lui demande la liste de mes taches.\n", leader)

		// on envoie un message pour demander la liste
		target := fmt.Sprintf("%s:%d",
			mapAdresse[leader],
			clusterState[leader].PortTcpDataTask,
		)

		conn, err := net.Dial("tcp", target)
		if err != nil {
			log.Println("Erreur connexion election choixLeader():", err)
			return
		}
		defer conn.Close()

		msg := TaskListDemandeur{
			ID_demandeur: config.Name,
		}

		data, err := json.Marshal(msg)
		if err != nil {
			log.Println("Erreur marshal:", err)
			return
		}

		env := common.Envelope{
			Type: TypeAskTasks,
			Data: data,
		}

		err = json.NewEncoder(conn).Encode(env)
		if err != nil {
			log.Println("Erreur envoi:", err)
			return
		}

		// Attente de la réponse
		var envRep common.Envelope
		if err = json.NewDecoder(conn).Decode(&envRep); err != nil {
			log.Println("Erreur lecture enveloppe réponse:", err)
			return
		}

		var rep MessageTaskList
		if err = json.Unmarshal(envRep.Data, &rep); err != nil {
			log.Println("Erreur décodage MessageTaskList:", err)
			return
		}

		handleRecvTasks(rep)

		log.Println("Fin de la migration de :", config.Name)

	}
}
