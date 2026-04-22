package main

import (
    "encoding/binary"
    "fmt"
    "net"
)

func ask_values() {
    conn, err := net.Dial("unix", "/tmp/cpu.sock")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    buf := make([]byte, 8+8) // uint64 + float64
	for {
		_, err = conn.Read(buf)
		if err != nil {
			panic(err)
		}
	
		mem := binary.LittleEndian.Uint64(buf[0:8])
		freq := mathFromBits(buf[8:16])
	
		fmt.Println("mem:", mem)
		fmt.Println("freq:", freq)
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