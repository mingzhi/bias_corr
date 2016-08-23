package main

import "math"

// Calculator is a correlation calculator.
type Calculator struct {
	Input         chan Pop
	Output        chan CorrResult
	LargeSize     int
	SmallSize     int
	ClusterNumber int
	MaxLen        int
	Repeat        int
	Bias          bool
}

// NewCalculator returns a new Calculator.
func NewCalculator() *Calculator {
	c := Calculator{}
	c.Input = make(chan Pop)
	c.Output = make(chan CorrResult)
	c.LargeSize = 10
	c.SmallSize = 1
	c.ClusterNumber = 1
	c.MaxLen = 100
	c.Repeat = 1
	c.Bias = true
	return &c
}

// Calculate calculate correlations.
func (c *Calculator) Calculate() {
	resChan := make(chan Result)
	go func() {
		defer close(resChan)
		for p := range c.Input {
			for k := 0; k < c.Repeat; k++ {
				genomes := choose(p, c.LargeSize, c.SmallSize, c.ClusterNumber, c.Bias)
				results := calcCorr(genomes, c.MaxLen)
				for _, r := range results {
					resChan <- r
				}
			}
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

func calcCorr(genomes []string, maxl int) (results []Result) {
	cms := calcCm(genomes, maxl)
	results = append(results, cms...)
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
