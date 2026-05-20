package main

import (
	"log"
	"os"
	"path/filepath"

	"runtime"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
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
	for {
		vm, err := mem.VirtualMemory()
		if err != nil {
			log.Println(err)
			continue
		}

		cpuPercent, err := cpu.Percent(
			time.Second,
			false,
		)

		if err != nil {
			log.Println(err)
			continue
		}

		availableCPU := int(
			float64(runtime.NumCPU()*100) -
				cpuPercent[0],
		)

		stateMu.Lock()

		state.Memory = int(vm.Available / 1024)

		state.CPU = availableCPU

		snapshot := state

		stateMu.Unlock()

		clusterState[config.Name] = snapshot

		if err := list.UpdateNode(
			2 * time.Second,
		); err != nil {
			log.Println("UpdateNode:", err)
		}

		time.Sleep(2 * time.Second)
	}
}
