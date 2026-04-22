package main

import (
	"encoding/json"

	"github.com/hashicorp/memberlist"
)

func startWorker(list *memberlist.Memberlist) {
	for {
		t := <-taskQueue
		var cpu = float64(^uint(0) >> 1) //
		var name = ""
		for nom, member := range clusterState {
			if member.CPU < cpu {
				cpu = member.CPU
				name = nom
			}
		}
		var target *memberlist.Node
		for _, n := range list.Members() {
			if n.Name == name {
				target = n
				break
			}
		}
		data, _ := json.Marshal(t)
		list.SendReliable(target, data)

	}
}
