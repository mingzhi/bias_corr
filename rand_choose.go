package main

import (
	"math/rand"
)

func randChooseClusters(p Pop, clusterSize int, num int) (clusters [][]string) {
	for i := 0; i < num; i++ {
		cluster := []string{}
		for k := 0; k < clusterSize; k++ {
			r := rand.Intn(p.Size)
			cluster = append(cluster, p.Genomes[r])
		}
		clusters = append(clusters, cluster)
	}
	return clusters
}
