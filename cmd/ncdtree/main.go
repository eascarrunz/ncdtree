package main

import (
	"bufio"
	"fmt"
	"ncdtree/pkg/fasta"
	"ncdtree/pkg/ncd"
	"ncdtree/pkg/phylocore"
	"os"
	"slices"

	"github.com/akamensky/argparse"
)

func main() {
	compressorList := []string{"Brotli", "Gzip"}

	parser := argparse.NewParser(
		"ncdtree",
		"Estimate a phylogeny from DNA sequences using the normalized compression distance (NCD) and neighbour-joining",
	)
	argInfile := parser.String(
		"f", "file",
		&argparse.Options{Required: false, Default: "", Help: "File with sequences in FASTA format"},
	)
	argAlgo := parser.Selector(
		"a", "algo",
		compressorList,
		&argparse.Options{Required: false, Default: "Brotli", Help: "Compression algorithm"},
	)
	argStats := parser.Flag(
		"s", "stats",
		&argparse.Options{Required: false, Help: "Print statistics"},
	)
	// argTag := parser.String(
	// 	"t", "tag",
	// 	&argparse.Options{Required: false, Default: "", Help: "Tag for names of output files"},
	// )
	argNoTree := parser.Flag(
		"", "notree",
		&argparse.Options{Required: false, Help: "Do estimate a tree. Only write out distance matrix."},
	)

	parser.Parse(os.Args)

	var input *os.File
	var err error
	var taxonNames *[]string
	var seqs *[][]byte

	if len(*argInfile) > 0 {
		input, err = os.Open(*argInfile)
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

	compressorName := *argAlgo

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

	if *argStats {
		selfNCD := make([]float64, N)
		var selfNCDMedian float64
		for i := range N {
			selfNCD[i] = ncd.NCD(cx[i], cx[i], cxx[i])
		}

		for i, taxonName := range *taxonNames {
			fmt.Printf("%3d. %-*s\t%d\t%.0f\t%.0f\t%f\n", i, 50, taxonName, len((*seqs)[i]), cx[i], cxx[i], selfNCD[i])
		}

		slices.Sort(selfNCD) // In-place sorting
		if N%2 == 0 {
			selfNCDMedian = (selfNCD[N/2] + selfNCD[(N/2)+1]) / 2
		} else {
			selfNCDMedian = selfNCD[(N+1)/2]
		}

		fmt.Printf("Median Self NCD: %f\n", selfNCDMedian)
	}

	// Create the distance matrix
	D := ncd.NCDMatrix(seqs, &cx, ctx)

	outFileMatrix, err := os.Create("ncd_matrix.txt")
	if err != nil {
		panic(err)
	}
	defer outFileMatrix.Close()
	ncd.WriteLabelledTriangularMatrix(outFileMatrix, taxonNames, D, 9)

	if !*argNoTree {
		taxset, err := phylocore.NewTaxonSet(*taxonNames)
		if err != nil {
			panic(err)
		}
		outFileTree, err := os.Create("tree.nwk")
		if err != nil {
			panic(err)
		}
		defer outFileTree.Close()
		tree := phylocore.NeighbourJoining(taxset, D)

		outFileTree.WriteString(tree.NewickString())
	}
}
