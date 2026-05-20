package main

import (
	"encoding/json"
	"log"

	"github.com/hashicorp/memberlist"
)

type MyDelegate struct{}

// les message statiques appelé par celui qui rejoin
func (d *MyDelegate) NodeMeta(limit int) []byte {
	stateMu.Lock()
	defer stateMu.Unlock()
	data, _ := json.Marshal(state) // convertir en json
	return data
}

// le message recu  (envoye par un autre noed)
func (d *MyDelegate) NotifyMsg(msg []byte) {

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
	if n.Name == config.Name {
		return
	}

	var s NodeState
	if err := json.Unmarshal(n.Meta, &s); err != nil {
		log.Println("json.Unmarshal error:", err)
		return
	}
	clusterState[n.Name] = s
	classifyNode(n.Name, s)
	mapAdresse[n.Name] = n.Addr.String()
}

// declahcé par moi quand qq un quit
func (e *MyEventDelegate) NotifyLeave(n *memberlist.Node) {
	delete(clusterState, n.Name)
	bucketMem.remove(n.Name)
	bucketCpu.remove(n.Name)
	bucketAvg.remove(n.Name)
	bucketLow.remove(n.Name)
	delete(mapAdresse, n.Name)
	log.Println("LEAVE:", n.Name)
}

// declaché pr moi quand le message recu par NodeMeta est different de l'ancien
func (e *MyEventDelegate) NotifyUpdate(n *memberlist.Node) {
	if n.Name == config.Name {
		return
	}

	var s NodeState
	if err := json.Unmarshal(n.Meta, &s); err != nil {
		log.Println("json.Unmarshal error:", err)
		return
	}
	clusterState[n.Name] = s
	classifyNode(n.Name, s)
	mapAdresse[n.Name] = n.Addr.String()
}

func classifyNode(name string, s NodeState) {
	// 1. On le supprime de TOUS les buckets par sécurité (O(1), très rapide)
	if name == config.Name {
		return
	}
	bucketMem.remove(name)
	bucketCpu.remove(name)
	bucketAvg.remove(name)
	bucketLow.remove(name)
	// 2. On le range dans le bon bucket
	if s.Memory >= 16000 {
		bucketMem.add(name)

	} else if s.CPU >= 80 {
		bucketCpu.add(name)

	} else if s.CPU >= 40 && s.Memory >= 4000 {
		bucketAvg.add(name)

	} else {
		bucketLow.add(name)
	}
}
