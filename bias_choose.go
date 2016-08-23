package main

import (
	"math/rand"
	"sort"
)

func choose(p Pop, largeSize, smallSize, clusterNum int, bias bool) (genomes []string) {
	if bias {
		return biasChoose(p, largeSize, smallSize, clusterNum)
	}

	return randomChoose(p, largeSize)
}

func randomChoose(p Pop, sampleSize int) (genomes []string) {
	for i := 0; i < sampleSize; i++ {
		index := rand.Intn(p.Size)
		genomes = append(genomes, p.Genomes[index])
	}

	return
}

func biasChoose(p Pop, largeSize, smallSize, clusterNum int) (genomes []string) {
	clusters := []int{}
	for i := 0; i < clusterNum; i++ {
		clusters = append(clusters, smallSize)
	}
	clusters = append(clusters, largeSize)

	indices := []int{}
	for k := 0; k < len(clusters); k++ {
		sampleSize := clusters[k]
		central := rand.Intn(len(p.Genomes))
		distances := calcDistances(p, central)
		tubles := make(Tubles, len(distances))
		for i := range distances {
			tubles[i] = Tuble{index: i, value: distances[i]}
		}
		sort.Sort(ByValue{tubles})

		for i := 0; i < sampleSize; i++ {
			index := tubles[i].index
			indices = append(indices, index)
		}
	}

	for _, i := range indices {
		genomes = append(genomes, p.Genomes[i])
	}

	return
}

func calcDistances(p Pop, i int) []float64 {
	distances := []float64{}
	for j := 0; j < len(p.Genomes); j++ {
		distances = append(distances, p.Ranks[i][j])
	}

	return distances
}

// Tuble stores index and value.
type Tuble struct {
	index int
	value float64
}

// Tubles is an array of Tuble
type Tubles []Tuble

// Len return the length of Tubles
func (s Tubles) Len() int { return len(s) }

// Swap swap the positions.
func (s Tubles) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByValue is a wrapper.
type ByValue struct{ Tubles }

// Less return true if i is less than j.
func (s ByValue) Less(i, j int) bool { return s.Tubles[i].value < s.Tubles[j].value }
