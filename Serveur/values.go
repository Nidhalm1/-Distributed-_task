package main

import (
    "encoding/binary"
    "fmt"
    "net"
	"os"
	"unsafe"
	"time"
)

func ask_values() {
	os.Remove("/tmp/cpu.sock")
    conn, err := net.Listen("unix", "/tmp/cpu.sock")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

	client, err := conn.Accept()
	if err != nil {
		panic(err)
	}
	defer client.Close()

    buf := make([]byte, 16) // uint64 + float64
	for {
		_, err = client.Read(buf)
		if err != nil {
			panic(err)
		}
	
		mem := binary.LittleEndian.Uint64(buf[0:8])
		freq := mathFromBits(buf[8:16])
	
		state.Memory = float64(mem)
		state.CPU = freq;
	}
}

// float64 helper
func mathFromBits(b []byte) float64 {
    bits := binary.LittleEndian.Uint64(b)
    return float64FromBits(bits)
}

func float64FromBits(b uint64) float64 {
    return *(*float64)(unsafe.Pointer(&b))
}

func print_values() {
	for {
		time.Sleep(2 * time.Second)
		fmt.Println(state.Memory)
		fmt.Println(state.CPU)
	}
}
