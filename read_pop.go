package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
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

func readPops(file string, max int) chan Pop {
	c := make(chan Pop, 20)
	go func() {
		defer close(c)
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		var decoder *json.Decoder
		appdix := file[len(file)-3:]
		if appdix == ".gz" {
			r, err := gzip.NewReader(f)
			if err != nil {
				panic(err)
			}
			defer r.Close()
			decoder = json.NewDecoder(r)
		} else {
			decoder = json.NewDecoder(f)
		}

		count := 0
		for count < max {
			var p Pop
			if err := decoder.Decode(&p); err != nil {
				if err != io.EOF {
					panic(err)
				}
				break
			}
			c <- p
			count++
		}
	}()
	return c
}
