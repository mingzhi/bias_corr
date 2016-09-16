package main

import "math"
import "runtime"

// Calculator is a correlation calculator.
type Calculator struct {
	Input      chan Pop
	Output     chan CorrResult
	Clusters   []int
	MaxLen     int
	Repeat     int
	GenomeLen  int
	Circular   bool
	ByCoalTime bool
	ByRandom   bool
	Mix        int
}

// NewCalculator returns a new Calculator.
func NewCalculator(clusters []int) *Calculator {
	c := Calculator{}
	c.Input = make(chan Pop)
	c.Output = make(chan CorrResult)
	c.Clusters = clusters
	c.MaxLen = 100
	c.Repeat = 1
	c.ByRandom = false
	c.Mix = 0
	return &c
}

// Calculate calculate correlations.
func (c *Calculator) Calculate() {
	resChan := make(chan Result)
	done := make(chan bool)
	ncpu := runtime.GOMAXPROCS(0)
	for i := 0; i < ncpu; i++ {
		go func() {
			for p := range c.Input {
				p2mvs := []*MeanVar{}
				var clusters [][]string
				if c.ByRandom {
					clusters = randChooseClusters(p, c.Clusters[0], c.Repeat)
				} else {
					clusters = biasChooseRank(p, c.Clusters[0], c.Repeat)
				}

				if c.Mix > 0 && !c.ByRandom {
					if c.Mix >= c.Clusters[0] {
						c.Mix = c.Clusters[0] - 1
					}
					mixes := randChooseClusters(p, c.Mix, c.Repeat)
					for k := 0; k < c.Repeat; k++ {
						for j := 1; j < len(clusters[k]); j++ {
							clusters[k][j] = mixes[k][j-1]
						}
					}
				}

				for k := 0; k < c.Repeat; k++ {
					genomes := clusters[k]
					if c.GenomeLen > 0 && c.GenomeLen < len(genomes[0]) {
						genomes = chopGenomes(genomes, c.GenomeLen)
						if c.MaxLen > c.GenomeLen {
							c.MaxLen = c.GenomeLen - 1
						}
					}
					results := calcCorr(genomes, c.MaxLen, c.Circular)
					for _, r := range results {
						resChan <- r
						if r.Type == "P2" {
							for len(p2mvs) <= r.Lag {
								p2mvs = append(p2mvs, NewMeanVar())
							}
							p2mvs[r.Lag].Add(r.Value)
						}
					}
				}
				ks := p2mvs[0].Mean()
				for l := 0; l < c.MaxLen; l++ {
					res := Result{}
					res.Lag = l
					res.Type = "Pn"
					res.N = p2mvs[l].N
					res.Value = p2mvs[l].Mean() / ks
					resChan <- res
				}
			}
			done <- true
		}()
	}

	go func() {
		defer close(resChan)
		for i := 0; i < ncpu; i++ {
			<-done
		}
	}()

	go func() {
		defer close(c.Output)
		resMap := collect(resChan, c.MaxLen)
		corrResults := getCorrResults(resMap)
		for _, cr := range corrResults {
			c.Output <- cr
		}
	}()

}

// chopGenomes
func chopGenomes(genomes []string, length int) []string {
	gs := []string{}
	if length > len(genomes[0]) {
		length = len(genomes[0])
	}
	for _, g := range genomes {
		gs = append(gs, g[:length])
	}
	return gs
}

// getCorrResults extract correlation results.
func getCorrResults(resMap map[string][]*MeanVar) []CorrResult {
	results := []CorrResult{}
	for t, mvs := range resMap {
		for i := 0; i < len(mvs); i++ {
			m := mvs[i].Mean()
			v := mvs[i].Variance()
			n := mvs[i].N
			c := CorrResult{L: i, M: m, V: v, N: n, T: t}
			results = append(results, c)
		}
	}
	return results
}

// CorrResult stores a correlation result.
type CorrResult struct {
	L int
	M float64
	V float64
	N int
	T string
	C int
}

func calcCorr(genomes []string, maxl int, circular bool) (results []Result) {
	p2s := calcP2(genomes, maxl, circular)
	results = append(results, p2s...)

	return
}

// collect averages correlation results.
func collect(resChan chan Result, maxLen int) map[string][]*MeanVar {
	resMap := make(map[string][]*MeanVar)
	for res := range resChan {
		for len(resMap[res.Type]) <= res.Lag {
			resMap[res.Type] = append(resMap[res.Type], NewMeanVar())
		}
		if !math.IsNaN(res.Value) {
			resMap[res.Type][res.Lag].Add(res.Value)
		}
	}

	return resMap
}
