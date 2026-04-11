package main

import (
	"log"
	"os"
	"strconv"

	"github.com/hashicorp/memberlist"
)

var config *memberlist.Config

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

	config.Delegate = &MyDelegate{}
	config.Events = &MyEventDelegate{} // cree un objet de Mydelegate qui sert pour quel info envoyé plus tard

	nullFile, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0) // redireger la sortie des logs vers fichier nul
	config.Logger = log.New(nullFile, "", 0)

	list, err := memberlist.Create(config) // démarre le protocole de communication entre les nodes et renvoie mon objet avec lequel je vais taffer
	if err != nil {
		log.Fatal(err)
	}
	if port != 7946 { // si je suis pas le premier je rejoins
		list.Join([]string{"127.0.0.1:7946"}) // essaie de rejoindre un cluster existant en se connectant à un node déjà présent
	}
	log.Println("Node:", config.Name, "started")
	list.Members()
	go boucle(list)
	select {}
}
