package phylocore

import (
	"bufio"
	"fmt"
	"ncdtree/pkg/ncd"
	"strconv"
	"strings"
)

func ReadDistanceMatrix(scanner bufio.Scanner) (*TaxonSet, *ncd.TriangularMatrix) {
	taxonNames := make([]string, 0)
	data := make([]float64, 0)

	i := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		taxonName := fields[0]
		taxonNames = append(taxonNames, taxonName)

		for j := 0; j < i; j += 1 {
			value, err := strconv.ParseFloat(fields[j+1], 64)
			if err != nil {
				panic(fmt.Errorf("invalid value %w", err))
			}
			data = append(data, value)
		}
		i += 1
	}

	nTaxa := i

	taxonSet, _ := NewTaxonSet(taxonNames)

	active := make([]bool, nTaxa)
	for i := range active {
		active[i] = true
	}
	m := &ncd.TriangularMatrix{N: nTaxa, RawData: data, Active: active}

	return taxonSet, m
}
