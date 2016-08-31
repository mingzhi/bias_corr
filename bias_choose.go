package main

import (
	"math/rand"
	"sort"
)

func biasChoose(p Pop, clusters []int, byCoalTime bool) (genomes []string) {
	indices := []int{}
	for k := 0; k < len(clusters); k++ {
		sampleSize := clusters[k]
		central := rand.Intn(len(p.Genomes))
		distances := calcDistances(p, central, byCoalTime)
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

func calcDistances(p Pop, i int, byCoalTime bool) []float64 {
	distances := []float64{}
	for j := 0; j < len(p.Genomes); j++ {
		if byCoalTime {
			distances = append(distances, p.Ranks[i][j])
		} else {
			distances = append(distances, compareGenomes(p.Genomes[i], p.Genomes[j]))
		}
	}

	return distances
}

func compareGenomes(a, b string) float64 {
	total := 0
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			total++
		}
	}
	return float64(total) / float64(len(a))
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
