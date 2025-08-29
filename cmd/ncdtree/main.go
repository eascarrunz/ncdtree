package main

import (
	"bufio"
	"fmt"
	"ncdtree/pkg/fasta"
	"ncdtree/pkg/ncd"
	"ncdtree/pkg/phylocore"
	"os"
	"slices"
)

func main() {
	var input *os.File
	var err error
	var taxonNames *[]string
	var seqs *[][]byte

	if len(os.Args) > 1 {
		input, err = os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		defer input.Close()
	} else {
		input = os.Stdin
	}

	scanner := bufio.NewScanner(input)
	taxonNames, seqs, err = fasta.ReadFasta(*scanner)
	if err != nil {
		panic(err)
	}

	N := len(*taxonNames)

	compressorName := "Brotli"

	fmt.Println("Compressor " + compressorName)

	var ctx *ncd.CompressionContext

	switch compressorName {
	case "Brotli":
		ctx = ncd.NewBrotliCompressionContext()
	case "Gzip":
		ctx = ncd.NewGzipCompressionContext()
	}

	cx := ncd.CXVector(seqs, *ctx)
	cxx := ncd.CXXVector(seqs, *ctx)
	selfNCD := make([]float64, N)

	for i := range N {
		selfNCD[i] = ncd.NCD(cx[i], cx[i], cxx[i])
	}

	for i, taxonName := range *taxonNames {
		fmt.Printf("%3d. %-*s\t%d\t%f\t%f\t%f\n", i, 50, taxonName, len((*seqs)[i]), cx[i], cxx[i], selfNCD[i])
	}

	slices.Sort(selfNCD) // In-place sorting
	var selfNCDMedian float64

	if N%2 == 0 {
		selfNCDMedian = (selfNCD[N/2] + selfNCD[(N/2)+1]) / 2
	} else {
		selfNCDMedian = selfNCD[(N+1)/2]
	}

	fmt.Printf("Median Self NCD: %f\n", selfNCDMedian)

	D := ncd.NCDMatrix(seqs, &cx, ctx)
	D.Show()

	taxset, err := phylocore.NewTaxonSet(*taxonNames)
	if err != nil {
		panic(err)
	}
	tree := phylocore.NeighbourJoining(taxset, D)

	fmt.Println(tree.NewickString())
}
