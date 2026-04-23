package main

import (
	"github.com/hashicorp/memberlist"
)

func startWorker(list *memberlist.Memberlist) {
	for {
		t := <-taskQueue
		if t.Estimatedmem > 8000 {

		}
	}
}
