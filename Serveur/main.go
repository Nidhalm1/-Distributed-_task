package main

import (
	"log"
	"net"
	"os"
	"strconv"

	"github.com/hashicorp/memberlist"
)

var config *memberlist.Config
var serverPort int

func main() {

	config = memberlist.DefaultLocalConfig()

	port := 7946
	serverPort = 1234
	addrJoin := ""

	// gossip port
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			port = p
		}
	}

	// tcp/grpc port
	if len(os.Args) > 2 {
		if p, err := strconv.Atoi(os.Args[2]); err == nil {
			serverPort = p
			state.PortTcp = p
		}
	}

	// node à rejoindre
	if len(os.Args) > 3 {
		addrJoin = os.Args[3]
	}

	config.Name = "node" + strconv.Itoa(port)

	config.BindPort = port
	config.AdvertisePort = port

	config.Delegate = &MyDelegate{}
	config.Events = &MyEventDelegate{}

	nullFile, _ := os.OpenFile(
		os.DevNull,
		os.O_WRONLY,
		0,
	)

	config.Logger = log.New(
		nullFile,
		"",
		0,
	)

	list, err := memberlist.Create(config)

	if err != nil {
		log.Fatal(err)
	}

	if addrJoin != "" {

		_, err := list.Join([]string{
			addrJoin,
		})

		if err != nil {
			log.Fatal(err)
		}

		log.Println(
			"Joined cluster via",
			addrJoin,
		)
	}

	log.Println(
		"Node:",
		config.BindPort,
		"started (tcp server on port",
		serverPort,
		")",
	)
	mapAdresse[config.Name] = config.AdvertiseAddr
	go startTCPServer(serverPort)
	go startWorker(list)
	go ask_values(list)

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
