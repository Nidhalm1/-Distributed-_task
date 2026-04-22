package main

import (
	"log"
	"net"
	"os"
	"strconv"

	"github.com/hashicorp/memberlist"
)

var config *memberlist.Config

func main() {

	config = memberlist.DefaultLocalConfig() // prepare la config du nord son nom , port ect
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
	// choose a separate TCP server port to avoid conflicting with memberlist's bind port
	serverPort := config.BindPort + 1
	// safety: keep port in valid range; if overflow, try decrementing instead
	if serverPort > 65535 {
		serverPort = config.BindPort - 1
	}
	log.Println("Node:", config.BindPort, "started (tcp server on port", serverPort, ")")
	// go boucle(list)
	go startTCPServer(serverPort)
	go startWorker(list)
	go ask_values()
	go print_values()
	select {}
}
func startTCPServer(serverPort int) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(serverPort))
	if err != nil {
		panic(err)
	}

	for {
		conn, _ := listener.Accept()
		go handleClient(conn)
	}
}
