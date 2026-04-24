package main

import (
	"NVPROJET/common"
	"fmt"
	"math/rand"

	"github.com/hashicorp/memberlist"
)

func startWorker(list *memberlist.Memberlist) {
	for {
		t := <-taskQueue
		var condidate []string
		if t.Estimatedmem >= 8000 {
			condidate = getIdeal(bucketMem, t)
		} else if t.Estimatedcpu >= 4000 {
			condidate = getIdeal(bucketCpu, t)
		} else if t.Estimatedcpu >= 2000 && t.Estimatedmem >= 2000 {
			condidate = getIdeal(bucketAvg, t)
		} else {
			condidate = getIdeal(bucketLow, t)
		}
		if len(condidate) > 0 {
			randomIdx := rand.Intn(len(condidate))
			chosenNode := condidate[randomIdx]
			fmt.Printf("Tâche assignée à : %s\n", chosenNode)
			// assigner la tâche à chosenNode ici

			continue
		} else { // len ==0
			fmt.Println("charge lourde pour ce system aucun node dispo pour elle , donc attendez")
			taskQueue <- t
			continue
		}
	}
}

func getIdeal(bucket *Bucket, t common.Task) []string {
	var maxTries = 5
	var condidate []string
	for range maxTries {
		var nodeName = bucket.pickRandom()
		if nodeName == "" {
			break
		}
		if clusterState[nodeName].Memory < t.Estimatedmem {
			continue
		}
		if clusterState[nodeName].CPU < t.Estimatedcpu {
			continue
		}
		exists := false
		for _, n := range condidate {
			if n == nodeName {
				exists = true
				break
			}
		}
		if !exists {
			condidate = append(condidate, nodeName)
		}
		if len(condidate) == 3 {
			break
		}
	}
	return condidate
}
