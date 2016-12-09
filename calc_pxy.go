package main

func calcPXY(ds1, ds2 []bool, maxl int, circular bool) []float64 {
	pxy := make([]float64, maxl)
	for l := 0; l < maxl; l++ {
		n := 0
		for i := 0; i < len(ds1); i++ {
			if !circular && i+l >= len(ds1) {
				break
			}
			x := ds1[i]
			y := ds2[(i+l)%len(ds1)]
			if !x && !y {
				pxy[l]++
			}
			n++
		}
		pxy[l] /= float64(n)
	}

	return pxy
}

func calcP2(genomes []string, maxl int, circular bool) (results []Result) {
	ds := make([]bool, len(genomes[0]))
	pxy := make([]float64, maxl)
	n := 0
	for i := 0; i < len(genomes); i++ {
		a := genomes[i]
		for j := i + 1; j < len(genomes); j++ {
			b := genomes[j]
			for k := 0; k < len(ds); k++ {
				ds[k] = a[k] == b[k]
			}
			xy := calcPXY(ds, ds, maxl, circular)
			for l := 0; l < maxl; l++ {
				pxy[l] += xy[l]
			}
			n++
		}
	}

	for l := 0; l < maxl; l++ {
		pxy[l] /= float64(n)
	}

	for i := 0; i < maxl; i++ {
		res := Result{}
		res.Lag = i
		res.N = n
		res.Type = "P2"
		res.Value = pxy[i]
		results = append(results, res)
	}

	return
}

func calcP3(genomes []string, maxl int, circular bool) (results []Result) {
	ds1 := make([]bool, len(genomes[0]))
	ds2 := make([]bool, len(genomes[0]))
	pxy := make([]float64, maxl)
	n := 0
	for i := 0; i < len(genomes); i++ {
		a := genomes[i]
		for j := 0; j < len(genomes); j++ {
			if i == j {
				continue
			}
			b := genomes[j]
			for k := 0; k < len(genomes); k++ {
				if k == i || k == j {
					continue
				}
				c := genomes[k]
				for w := 0; w < len(ds1); w++ {
					ds1[w] = a[w] == b[w]
					ds2[w] = a[w] == c[w]
				}
				xy := calcPXY(ds1, ds2, maxl, circular)
				for l := 0; l < maxl; l++ {
					pxy[l] += xy[l]
				}
				n++
			}
		}
	}

	for l := 0; l < maxl; l++ {
		pxy[l] /= float64(n)
	}

	for i := 0; i < maxl; i++ {
		res := Result{}
		res.Lag = i
		res.N = n
		res.Type = "P3"
		res.Value = pxy[i]
		results = append(results, res)
	}

	return
}

func calcP4(genomes []string, maxl int, circular bool) (results []Result) {
	ds1 := make([]bool, len(genomes[0]))
	ds2 := make([]bool, len(genomes[0]))
	pxy := make([]float64, maxl)
	n := 0
	for i := 0; i < len(genomes); i++ {
		a := genomes[i]
		for j := 0; j < len(genomes); j++ {
			if i == j {
				continue
			}
			b := genomes[j]
			for k := 0; k < len(genomes); k++ {
				if k == i || k == j {
					continue
				}
				c := genomes[k]
				for h := 0; h < len(genomes); h++ {
					if h == i || h == j || h == k {
						continue
					}
					d := genomes[h]
					for w := 0; w < len(ds1); w++ {
						ds1[w] = a[w] == b[w]
						ds2[w] = c[w] == d[w]
					}
					xy := calcPXY(ds1, ds2, maxl, circular)
					for l := 0; l < maxl; l++ {
						pxy[l] += xy[l]
					}
					n++
				}
			}
		}
	}

	for l := 0; l < maxl; l++ {
		pxy[l] /= float64(n)
	}

	for i := 0; i < maxl; i++ {
		res := Result{}
		res.Lag = i
		res.N = n
		res.Type = "P4"
		res.Value = pxy[i]
		results = append(results, res)
	}

	return
}
