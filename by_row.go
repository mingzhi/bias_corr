package main

import (
	"math"
	"sort"

	"bitbucket.org/mingzhi/seqcorr/nuclcov"
)

type Result struct {
	Lag   int
	Value float64
	N     int
	Type  string
}

// Sub records the position and
type Sub struct {
	Pos int
	A   byte
}

// Subs is a list of Sub.
type Subs []Sub

// Len returns the length of Subs.
func (s Subs) Len() int { return len(s) }

// Swap swap the values at two positions.
func (s Subs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// BySubPos is a wrapper for sorting.
type BySubPos struct{ Subs }

// Less return true if the value at i is less than the value at j.
func (s BySubPos) Less(i, j int) bool {
	return s.Subs[i].Pos < s.Subs[j].Pos
}

func calcCs(genomes []string, maxl int) (results []Result) {
	matrix := [][]*nuclcov.NuclCov{}
	for _, genome := range genomes {
		for i := 0; i < len(genome); i++ {
			for j := i; j < len(genome) && j-i < maxl; j++ {
				pos := i
				lag := j - i
				a := genome[i]
				b := genome[j]
				for len(matrix) <= pos {
					matrix = append(matrix, []*nuclcov.NuclCov{})
				}

				for len(matrix[pos]) <= lag {
					matrix[pos] = append(matrix[pos], nuclcov.New([]byte{'1', '2', '3', '4'}))
				}

				matrix[pos][lag].Add(a, b)
			}
		}
	}

	for lag := 0; lag < maxl; lag++ {
		mc := NewMeanCov()
		for i := 0; i < len(matrix); i++ {
			if lag < len(matrix[i]) {
				xy, xbar, ybar, n := matrix[i][lag].Cov()
				if !math.IsNaN(xy) {
					mc.Add(xy, xbar, ybar, n)
				}
			}
		}

		cs := mc.Mean.GetResult()
		cr := mc.Cov.GetResult()
		p2 := mc.MeanXY()
		n := mc.Mean.GetN()

		crRes := Result{Value: cr, Lag: lag, N: n, Type: "Cr"}
		csRes := Result{Value: cs, Lag: lag, N: n, Type: "Cs"}
		p2Res := Result{Value: p2, Lag: lag, N: n, Type: "P2"}

		results = append(results, []Result{crRes, csRes, p2Res}...)
	}

	return
}

func calcCm(genomes []string, maxl int) (results []Result) {
	cm := make([]float64, maxl)
	p2 := make([]float64, maxl)
	d := 0.0
	vd := 0.0

	xy := make([]float64, maxl)
	diff := make([]bool, len(genomes[0]))
	for i := range genomes {
		a := genomes[i]
		for j := i + 1; j < len(genomes); j++ {
			b := genomes[j]
			for k := 0; k < len(a); k++ {
				diff[k] = a[k] == b[k]
			}

			var xbar, ybar float64
			for l := 0; l < maxl; l++ {
				xy[l] = 0
				for k := 0; k < len(a); k++ {
					x := diff[k]
					y := diff[(k+l)%len(a)]
					if x == false && y == false {
						xy[l]++
					}
				}

				if l == 0 {
					xbar = xy[l] / float64(len(a))
					ybar = xbar
					d += xbar
					vd += xbar * ybar
				}

				v := xy[l] / float64(len(a))
				cm[l] += v - xbar*ybar
				p2[l] += v
			}

		}
	}

	n := len(genomes) * (len(genomes) - 1) / 2

	for i := 0; i < maxl; i++ {
		res := Result{}
		res.Lag = i
		res.N = n
		res.Type = "Cm"
		res.Value = cm[i] / float64(n)
		results = append(results, res)
	}

	ks := d / float64(n)
	vard := vd/float64(n) - ks*ks

	results = append(results, Result{Lag: 0, N: n, Type: "Ks", Value: ks})
	results = append(results, Result{Lag: 0, N: n, Type: "Vd", Value: vard})

	for i := 0; i < maxl; i++ {
		res := Result{}
		res.Lag = i
		res.N = n
		res.Type = "P2"
		res.Value = p2[i] / float64(n)
		results = append(results, res)
	}

	if ks > 0 {
		for i := 0; i < maxl; i++ {
			res := Result{}
			res.Lag = i
			res.N = n
			res.Type = "PN"
			res.Value = p2[i] / float64(n) / ks
			results = append(results, res)
		}
	}

	return
}

func identifySubs(genomes []string) (subsArr []Subs) {
	ref := genomes[0]
	subsArr = append(subsArr, Subs{})
	for i := 1; i < len(genomes); i++ {
		subs := Subs{}
		for k := 0; k < len(ref); k++ {
			if ref[k] != genomes[i][k] {
				subs = append(subs, Sub{Pos: k, A: genomes[i][k]})
			}
		}
		subsArr = append(subsArr, subs)
	}

	return
}

// removeDuplicateSubs
func removeDuplicateSubs(subs1 Subs, others ...Subs) Subs {
	allSubs := Subs{}
	allSubs = append(allSubs, subs1...)
	for _, subs := range others {
		allSubs = append(allSubs, subs...)
	}
	if len(allSubs) <= 1 {
		return allSubs
	}
	// remove same subsistutions.
	sort.Sort(BySubPos{allSubs})

	dedupSubs := Subs{}
	old := Sub{Pos: allSubs[0].Pos - 1, A: ' '}
	for _, s := range allSubs {
		if s.Pos == old.Pos {
			if s.A == old.A {
				dedupSubs = dedupSubs[:len(dedupSubs)-1]
			}
		} else {
			dedupSubs = append(dedupSubs, s)
		}
		old = s
	}

	return dedupSubs
}

// removeDuplicateInts
func removeDuplicateInts(values []int, others ...[]int) []int {
	all := []int{}
	others = append(others, values)
	for _, vs := range others {
		all = append(all, vs...)
	}

	if len(all) <= 1 {
		return all
	}

	sort.Ints(all)
	dedupInts := []int{}
	old := all[0] - 1
	for _, s := range all {
		if s == old {
			dedupInts = dedupInts[:len(dedupInts)-1]
		} else {
			dedupInts = append(dedupInts, s)
		}
		old = s
	}

	return dedupInts
}
