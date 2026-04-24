package main

import (
	"NVPROJET/common"
	"encoding/json"
	"log"
	"os"

	"os/exec"

	"github.com/hashicorp/memberlist"
)

type MyDelegate struct{}

// les message statiques
func (d *MyDelegate) NodeMeta(limit int) []byte {
	data, _ := json.Marshal(state) // convertir en json
	return data
}

// le message recu  (envoye par un autre noed)
func (d *MyDelegate) NotifyMsg(msg []byte) {
	var t common.Task
	if err := json.Unmarshal(msg, &t); err != nil {
		log.Println("json.Unmarshal error:", err)
		return
	} // rasjouter focntion handle  au lieu deecrire direct
	cmd := exec.Command(t.Command, t.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("command execution error: %v", err)
	}
}

// message à envoyer quand on veut à tt le monde
func (d *MyDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

// appeler par moi pour celui qui me contact pr rejoindre pr lui envoyer
func (d *MyDelegate) LocalState(join bool) []byte {
	data, _ := json.Marshal(clusterState)
	return data
}

// appelé chez le noeud qui rejoins pr recevoir le mssg de LocalState et
func (d *MyDelegate) MergeRemoteState(buf []byte, join bool) {
	var recu map[string]NodeState
	if err := json.Unmarshal(buf, &recu); err != nil {
		log.Println("json.Unmarshal error:", err)
		return
	}
	for name, state := range recu {
		clusterState[name] = state
		classifyNode(name, state)
	}
}

type MyEventDelegate struct{}

// declahcé par moi quand qq un join
func (e *MyEventDelegate) NotifyJoin(n *memberlist.Node) {
	var s NodeState
	if err := json.Unmarshal(n.Meta, &s); err != nil {
		log.Println("json.Unmarshal error:", err)
		return
	}
	clusterState[n.Name] = s
	classifyNode(n.Name, s)
}

// declahcé par moi quand qq un quit
func (e *MyEventDelegate) NotifyLeave(n *memberlist.Node) {
	delete(clusterState, n.Name)
	bucketMem.remove(n.Name)
	bucketCpu.remove(n.Name)
	bucketAvg.remove(n.Name)
	bucketLow.remove(n.Name)
	log.Println("LEAVE:", n.Name)
}

// declaché pr moi quand le message recu par NodeMeta est different de l'ancien
func (e *MyEventDelegate) NotifyUpdate(n *memberlist.Node) {
	var s NodeState
	if err := json.Unmarshal(n.Meta, &s); err != nil {
		log.Println("json.Unmarshal error:", err)
		return
	}
	clusterState[n.Name] = s
	classifyNode(n.Name, s)
	log.Println("UPDATE:", n.Name)
}

func classifyNode(name string, s NodeState) {
	// 1. On le supprime de TOUS les buckets par sécurité (O(1), très rapide)
	bucketMem.remove(name)
	bucketCpu.remove(name)
	bucketAvg.remove(name)
	bucketLow.remove(name)

	// 2. On le range dans le bon bucket
	if s.Memory >= 8000 {
		bucketMem.add(name)
	} else if s.CPU >= 4000 {
		bucketCpu.add(name)
	} else if s.CPU >= 2000 && s.Memory >= 2000 {
		bucketAvg.add(name)
	} else {
		bucketLow.add(name)
	}
}
