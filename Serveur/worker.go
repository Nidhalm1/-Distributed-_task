package main

import (
	"NVPROJET/common"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

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
			ctx, cancel := context.WithCancel(context.Background())
			winnerChan := make(chan string, 1)
			for _, node := range condidate {
				go func(n string) {
					var addr = mapAdresse[n]
					var port = clusterState[n].PortTcp
					conn, err := net.DialTimeout("tcp",
						net.JoinHostPort(addr, fmt.Sprintf("%d", port)),
						200*time.Millisecond,
					)
					if err != nil {
						fmt.Println("pas recu à se connecter au noeud aleatoire")

						return
					}

					go func() {
						<-ctx.Done()
						conn.Close()
					}()

					encoder := json.NewEncoder(conn)
					decoder := json.NewDecoder(conn)

					var probe = common.Probe{Estimatedmem: t.Estimatedmem, Estimatedcpu: t.Estimatedcpu}
					encoder.Encode(probe)
					var probeRep common.ProbeResponse
					if err := decoder.Decode(&probeRep); err != nil {
						return
					}
					if probeRep.Accepted {
						//arreter les decodes des autres
						select {
						case winnerChan <- n:
							cancel() // stop les autres
						default:
						}
					}
				}(node)
			}
			select { // att jusqu'a un de ses evenemtn
			case chosen := <-winnerChan:
				cancel()
				fmt.Println("Node choisi:", chosen)
				var addr = mapAdresse[chosen]
				var port = clusterState[chosen].PortTcp
				conn, err := net.DialTimeout("tcp",
					net.JoinHostPort(addr, fmt.Sprintf("%d", port)),
					200*time.Millisecond,
				)
				if err != nil {
					fmt.Println("échec connexion au node choisi")
					continue
				}
				encoder := json.NewEncoder(conn)
				decoder := json.NewDecoder(conn)
				encoder.Encode(t)
				var taskResult common.TaskResult
				decoder.Decode(&taskResult)
				conn.Close()
				tasks[t.ID] = taskResult
				continue
			case <-time.After(300 * time.Millisecond):
				fmt.Println("Aucun node dispo")
				cancel()
				go func() {
					taskQueue <- t
				}()
				fmt.Println("je retente plus tard")
			}
		} else { // len ==0
			fmt.Println("charge lourde pour ce system aucun node dispo pour elle , donc attendez")
			go func() {
				taskQueue <- t
			}()
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
