package main

import (
	"bufio"
	"fmt"
	"ncdtree/pkg/fasta"
	"ncdtree/pkg/ncd"
	"ncdtree/pkg/phylocore"
	"os"
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

	// ctx := ncd.NewGzipCompressionContext()
	// ctx := ncd.NewBrotliCompressionContext()
	fmt.Println("LZMA2 Compressor")
	ctx := ncd.NewLZMACompressionContext()

	cx := ncd.CXVector(seqs, *ctx)
	cxx := ncd.CXXVector(seqs, *ctx)
	cratio := make([]float64, N)

	for i := range N {
		cratio[i] = cx[i] / cxx[i]
	}

	for i, taxonName := range *taxonNames {
		fmt.Printf("%3d. %-*s\t%d\t%f\t%f\t%f\n", i, 50, taxonName, len((*seqs)[i]), cx[i], cxx[i], cratio[i])
	}

	D := ncd.NCDMatrix(seqs, &cx, ctx)
	D.Show()

	taxset, err := phylocore.NewTaxonSet(*taxonNames)
	if err != nil {
		panic(err)
	}
	tree := phylocore.NeighbourJoining(taxset, D)

	fmt.Println(tree.NewickString())
}
