import (
    "fmt"
    "net"
    "os"
)

func ask_values() {
	os.Remove("/tmp/cpu.sock")

    ln, err := net.Listen("unix", "/tmp/cpu.sock")
    if err != nil {
        panic(err)
    }
    defer ln.Close()

    fmt.Println("Unix socket listening")

    for {
        conn, _ := ln.Accept()
        go handle(conn)
    }
}

func handle(conn net.Conn) {
    defer conn.Close()

    buf := make([]byte, 1024)

    for {
        n, err := conn.Read(buf)
        if err != nil {
            return
        }

        
    }
}
