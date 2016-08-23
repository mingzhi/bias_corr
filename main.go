package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/alecthomas/kingpin"
)

// Pop stores a population simulation results.
type Pop struct {
	Size, Length               int
	MutationRate, TransferRate float64
	FragLen                    int
	Generation                 int
	Genomes                    []string
	Ranks                      [][]float64
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

func readPops(file string) chan Pop {
	c := make(chan Pop, 20)
	go func() {
		defer close(c)
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		r, err := gzip.NewReader(f)
		if err != nil {
			panic(err)
		}
		defer r.Close()
		count := 0
		decoder := json.NewDecoder(r)
		for {
			var p Pop
			if err := decoder.Decode(&p); err != nil {
				if err != io.EOF {
					panic(err)
				}
				break
			}
			c <- p
			count++
			fmt.Println(count)
		}
	}()
	return c
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
