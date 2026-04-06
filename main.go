package main

import (
	"log"
	"os"
	"strconv"

	"encoding/json"

	"github.com/hashicorp/memberlist"
)

type MyDelegate struct{}

// les messageà envoyer statique et automatiquement
func (d *MyDelegate) NodeMeta(limit int) []byte {
	meta := map[string]string{
		"load":   "5",
		"cpu":    "20%",
		"memory": "10%",
	}
	data, _ := json.Marshal(meta) // convertir en json
	return data
}

// le message recu par (envoye par un autre noed)
func (d *MyDelegate) NotifyMsg(msg []byte) {}

// message à envoyer quand on veut à tt le monde
func (d *MyDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

// appeler seuelet lorsuqeu un noed rejoind pr lui envoyer l'etat complet
func (d *MyDelegate) LocalState(join bool) []byte {
	return nil
}

// appelé chez le noeud qui rejoins l'etat
func (d *MyDelegate) MergeRemoteState(buf []byte, join bool) {}

func main() {

	config := memberlist.DefaultLocalConfig() // prepare la config du nord son nom , port ect
	// Récupère le port depuis les arguments du programme
	port := 7946 // valeur par défaut
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			port = p
		}
	}
	config.Name = "node" + strconv.Itoa(port)
	config.BindPort = port
	config.AdvertisePort = port
	config.Delegate = &MyDelegate{}                        // cree un objet de Mydelegate qui sert pour quel info envoyé plus tard
	nullFile, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0) // redireger la sortie des logs vers fichier nul
	config.Logger = log.New(nullFile, "", 0)
	list, err := memberlist.Create(config) // démarre le protocole de communication entre les nodes
	if err != nil {
		log.Fatal(err)
	}
	if port != 7946 { // si je sois pas le premier je rejoins
		list.Join([]string{"127.0.0.1:7946"}) // essaie de rejoindre un cluster existant en se connectant à un node déjà présent
	}
	log.Println("Node:", config.Name, "started")

	select {}
}
