package main

import "math/rand"

var bucketMem *Bucket
var bucketCpu *Bucket
var bucketAvg *Bucket
var bucketLow *Bucket

func init() {
	bucketMem = createBucket()
	bucketCpu = createBucket()
	bucketAvg = createBucket()
	bucketLow = createBucket()
}

type Bucket struct {
	nodes    []string
	indexMap map[string]int
}

func (b *Bucket) remove(node string) {
	idx, exists := b.indexMap[node]
	if !exists {
		return
	}

	lastIndex := len(b.nodes) - 1
	lastNode := b.nodes[lastIndex]

	// swap
	b.nodes[idx] = lastNode
	b.indexMap[lastNode] = idx
	b.nodes = b.nodes[:lastIndex]

	delete(b.indexMap, node)
}

func (b *Bucket) add(node string) {
	if _, exists := b.indexMap[node]; exists {
		return
	}
	b.nodes = append(b.nodes, node)
	var idx = len(b.nodes) - 1
	b.indexMap[node] = idx
}

func (b *Bucket) pickRandom() string {
	if len(b.nodes) == 0 {
		return ""
	}
	var r = rand.Intn(len(b.nodes))
	return b.nodes[r]
}

func createBucket() *Bucket {
	b := &Bucket{
		nodes:    make([]string, 0), //liste vide
		indexMap: make(map[string]int),
	}
	return b
}
