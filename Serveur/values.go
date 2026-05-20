package main

import (
	"encoding/binary"
	"io"
	"log"
	"math"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hashicorp/memberlist"
)

func sockPath() string {
	if p := os.Getenv("CPU_SOCK_PATH"); p != "" {
		return p
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(os.TempDir(), "cpu.sock") // ex: C:\Users\...\AppData\Local\Temp\cpu.sock
	}
	return "/tmp/cpu.sock"
}

func ask_values(list *memberlist.Memberlist) {
	path := sockPath()
	os.Remove(path)
	listener, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal("ask_values listen:", err)
	}
	defer listener.Close()
	log.Println(">>> ask_values: listening on", path) // <-- AJOUTE CA
	client, err := listener.Accept()
	if err != nil {
		log.Fatal("ask_values accept:", err)
	}
	defer client.Close()
	log.Println(">>> ask_values: collector connecté") // <-- ET CA

	buf := make([]byte, 16)
	for {
		if _, err := io.ReadFull(client, buf); err != nil {
			log.Println("ask_values read:", err)
			return
		}

		mem := binary.LittleEndian.Uint64(buf[0:8]) // en kB
		freq := math.Float64frombits(binary.LittleEndian.Uint64(buf[8:16]))
		log.Printf("Memory: %d kB, CPU Frequency: %.2f MHz\n", mem, freq)

		stateMu.Lock()
		state.Memory = int(mem) // kB — voir note ci-dessous
		state.CPU = int(freq)   // MHz dispo
		snapshot := state
		stateMu.Unlock()

		// reclassifier le nœud local (memberlist n'appelle pas NotifyUpdate sur soi)
		clusterState[config.Name] = snapshot
		// déclencher le rebroadcast du Meta vers les autres nœuds
		if err := list.UpdateNode(2 * time.Second); err != nil {
			log.Println("UpdateNode:", err)
		}
	}

}
