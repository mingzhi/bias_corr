package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"log"
	"strconv"
	"strings"

	"runtime"

	"github.com/alecthomas/kingpin"
	"github.com/cheggaaa/pb"
)

func main() {
	input := kingpin.Flag("input", "input population simulation results").Required().String()
	output := kingpin.Flag("output", "output").Required().String()
	clusterStr := kingpin.Flag("clusters", "clusters").Required().String()
	numPop := kingpin.Flag("num_pop", "number of populations").Required().Int()
	maxLen := kingpin.Flag("maxl", "max len of correlations").Default("100").Int()
	repeat := kingpin.Flag("repeat", "repeat").Default("10").Int()
	showProgress := kingpin.Flag("progress", "show progress").Default("false").Bool()
	genomeLen := kingpin.Flag("genome_length", "genome length").Default("0").Int()
	circularGenome := kingpin.Flag("circular_genome", "circular genome").Default("false").Bool()
	ncpu := kingpin.Flag("ncpu", "number of CPUs for using").Default("0").Int()
	byCoalTime := kingpin.Flag("by_coal_time", "compare genome by coalescent time").Default("false").Bool()
	byRandom := kingpin.Flag("by_random", "choose clusters by random").Default("false").Bool()

	kingpin.Parse()
	rand.Seed(time.Now().UTC().UnixNano())

	if *ncpu == 0 {
		*ncpu = runtime.NumCPU()
		runtime.GOMAXPROCS(*ncpu)
	}

	clusters := getClusters(*clusterStr)
	c := NewCalculator(clusters)
	c.MaxLen = *maxLen
	c.Repeat = *repeat
	c.GenomeLen = *genomeLen
	c.Circular = *circularGenome
	c.ByCoalTime = *byCoalTime
	c.ByRandom = *byRandom

	popChan := readPops(*input, *numPop)
	go func() {
		defer close(c.Input)

		var bar *pb.ProgressBar
		if *showProgress {
			bar = pb.StartNew(*numPop)
			defer bar.Finish()
		}

		for pop := range popChan {
			c.Input <- pop
			if *showProgress {
				bar.Increment()
			}
		}
	}()

	c.Calculate()
	write(c.Output, *output)
}

func getClusters(s string) []int {
	terms := strings.Split(s, ",")
	clusters := []int{}
	for i := range terms {
		v, err := strconv.Atoi(terms[i])
		if err != nil {
			log.Panicf("Error when convert %s to integer: %v", terms[i], err)
		}
		clusters = append(clusters, v)
	}
	return clusters
}

// write the final result.
func write(results chan CorrResult, outFile string) {
	w, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	w.WriteString("l,m,v,n,t\n")
	for res := range results {
		n := res.N
		m := res.M
		v := res.V
		i := res.L
		t := res.T
		if n > 0 && !math.IsNaN(v) {
			w.WriteString(fmt.Sprintf("%d", i))
			w.WriteString(fmt.Sprintf(",%g,%g", m, v))
			w.WriteString(fmt.Sprintf(",%d,%s\n", n, t))
		}
	}
}
