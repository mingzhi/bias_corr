package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"io"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"
)

// Pop stores a population simulation results.
type Pop struct {
	Size, Length               int
	MutationRate, TransferRate float64
	FragLen                    int
	Generation                 int
	Genomes                    []string
}

func main() {
	input := kingpin.Flag("input", "input population simulation results").Required().String()
	output := kingpin.Flag("output", "output").Required().String()
	smallSize := kingpin.Flag("small_cluster", "small cluster size").Default("5").Int()
	largeSize := kingpin.Flag("large_cluster", "large cluster size").Default("50").Int()
	clusterNum := kingpin.Flag("num_cluster", "number of clusters").Default("1").Int()
	ncpu := kingpin.Flag("ncpu", "number of cpus").Default("1").Int()
	maxLen := kingpin.Flag("maxl", "max len of correlations").Default("100").Int()
	bias := kingpin.Flag("bias", "bias sampling").Default("false").Bool()
	repeat := kingpin.Flag("repeat", "repeat").Default("10").Int()

	kingpin.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	popChan := readPops(*input)
	resChan := make(chan Result, 20)
	done := make(chan bool)
	for i := 0; i < *ncpu; i++ {
		go func() {
			for p := range popChan {
				for k := 0; k < *repeat; k++ {
					genomes := choose(p, *largeSize, *smallSize, *clusterNum, *bias)
					results := calcCorr(genomes, *maxLen)
					for _, r := range results {
						resChan <- r
					}
				}
			}
			done <- true
		}()
	}

	go func() {
		defer close(resChan)
		for i := 0; i < *ncpu; i++ {
			<-done
		}
	}()

	results := collect(resChan, *maxLen)
	write(results, *output)

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

func readPops(file string) chan Pop {
	c := make(chan Pop, 20)
	go func() {
		defer close(c)
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		decoder := json.NewDecoder(f)
		for {
			var p Pop
			if err := decoder.Decode(&p); err != nil {
				if err != io.EOF {
					panic(err)
				}
				break
			}
			c <- p
		}
	}()
	return c
}

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
		d := 0.0
		if i != j {
			d = calcDist(p.Genomes[i], p.Genomes[j])
		}
		distances = append(distances, d)
	}

	return distances
}

func calcDist(a, b string) float64 {
	d := 0
	for i := range a {
		if a[i] != b[i] {
			d++
		}
	}
	return float64(d) / float64(len(a))
}

func calcCorr(genomes []string, maxl int) (results []Result) {
	cms := calcCmSub(genomes, maxl)
	css := calcCs(genomes, maxl)
	results = append(results, cms...)
	results = append(results, css...)

	cm := make([]float64, maxl)
	cs := make([]float64, maxl)
	ns := make([]int, maxl)
	for _, c := range cms {
		if c.Type == "Cm" && c.Lag < maxl {
			cm[c.Lag] = c.Value
			ns[c.Lag] = c.N
		}
	}

	for _, c := range css {
		if c.Type == "Cs" && c.Lag < maxl {
			cs[c.Lag] = c.Value
		}
	}

	for i := 0; i < len(cm); i++ {
		r := Result{}
		r.Lag = i
		r.N = ns[i]
		r.Type = "Ct"
		r.Value = cm[i] / cs[i]
		if !math.IsInf(r.Value, 0) && !math.IsNaN(r.Value) {
			results = append(results, r)
		}
	}

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

// write the final result.
func write(result map[string][]*MeanVar, outFile string) {
	w, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	w.WriteString("l,m,v,n,t\n")
	for t, mvs := range result {
		for i := 0; i < len(mvs); i++ {
			m := mvs[i].Mean()
			v := mvs[i].Variance()
			n := mvs[i].N
			if n > 0 && !math.IsNaN(v) {
				w.WriteString(fmt.Sprintf("%d", i))
				w.WriteString(fmt.Sprintf(",%g,%g", m, v))
				w.WriteString(fmt.Sprintf(",%d,%s\n", n, t))
			}
		}
	}
}

// MeanVar is for calculate mean and variance in the increment way.
type MeanVar struct {
	N             int     // number of values.
	M1            float64 // first moment.
	Dev           float64
	NDev          float64
	M2            float64 // second moment.
	BiasCorrected bool
}

// NewMeanVar return a new MeanVar.
func NewMeanVar() *MeanVar {
	return &MeanVar{}
}

// Add adds a value.
func (m *MeanVar) Add(v float64) {
	if m.N < 1 {
		m.M1 = 0
		m.M2 = 0
	}

	m.N++
	n0 := m.N
	m.Dev = v - m.M1
	m.NDev = m.Dev / float64(n0)
	m.M1 += m.NDev
	m.M2 += float64(m.N-1) * m.Dev * m.NDev
}

// Mean returns the mean result.
func (m *MeanVar) Mean() float64 {
	return m.M1
}

// Variance returns the variance.
func (m *MeanVar) Variance() float64 {
	if m.N < 2 {
		return math.NaN()
	}

	if m.BiasCorrected {
		return m.M2 / float64(m.N-1)
	}

	return m.M2 / float64(m.N)
}

// Append add another result.
func (m *MeanVar) Append(m2 *MeanVar) {
	if m.N == 0 {
		m.N = m2.N
		m.M1 = m2.M1
		m.Dev = m2.Dev
		m.NDev = m2.NDev
		m.M2 = m2.M2
	} else {
		if m2.N > 0 {
			total1 := m.M1 * float64(m.N)
			total2 := m2.M1 * float64(m2.N)
			newMean := (total1 + total2) / float64(m.N+m2.N)
			delta1 := m.Mean() - newMean
			delta2 := m2.Mean() - newMean
			sm := (m.M2 + m2.M2) + float64(m.N)*delta1*delta1 + float64(m2.N)*delta2*delta2
			m.M1 = newMean
			m.M2 = sm
			m.N = m.N + m2.N
		}
	}
}
