package main

import (
	"bufio"
	"fmt"
	"ncdtree/pkg/fasta"
	"ncdtree/pkg/ncd"
	"ncdtree/pkg/phylocore"
	"os"
	"strings"

	"github.com/akamensky/argparse"
	"github.com/google/brotli/go/cbrotli"
	"github.com/klauspost/compress/zstd"
)

const inputBufSize = 64 * 1024

func main() {
	compressorList := []string{"Brotli", "Gzip", "zstd"}

	parser := argparse.NewParser(
		"ncdtree",
		"Estimate a phylogeny from DNA sequences using the normalized compression distance (NCD) and neighbour-joining",
	)
	argInfile := parser.String(
		"f", "file",
		&argparse.Options{Required: false, Help: "File with sequences in FASTA format (read from stdin if none is given)"},
	)
	argAlgo := parser.Selector(
		"Z", "compressor",
		compressorList,
		&argparse.Options{Required: false, Default: "Brotli", Help: "Compression algorithm"},
	)
	argZCmd := parser.String(
		"c", "cmd",
		&argparse.Options{Required: false, Default: "", Help: "External compressor command (overrides -Z). Must read from stdin and write compressed data to stdout, e.g. 'gzip -c'."},
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
		&argparse.Options{Required: false, Help: "Do not estimate a tree. Only write out distance matrix."},
	)

	parser.Parse(os.Args)

	var input *os.File
	var err error
	var taxonNames *[]string
	var seqs *[][]byte
	var inputStat os.FileInfo

	if len(*argInfile) > 0 {
		input, err = os.Open(*argInfile)
		if err != nil {
			panic(err)
		}
		inputStat, err = input.Stat()
		if err != nil {
			panic(err)
		}
		if inputStat.Size() == 0 {
			os.Stderr.WriteString("Empty input file.\n")
			os.Exit(65)
		}
		defer input.Close()
	} else {
		input = os.Stdin
		inputStat, err = input.Stat()
		if err != nil {
			panic(err)
		}
		if inputStat.Mode()&os.ModeCharDevice != 0 {
			os.Stderr.WriteString("No input.\n")
			os.Exit(66)
		}
	}

	reader := bufio.NewReaderSize(input, inputBufSize)
	taxonNames, seqs, err = fasta.ReadFasta(reader)
	if err != nil {
		panic(err)
	}

	N := len(*taxonNames)

	compressorName := *argAlgo

	var mc ncd.ManagedCompressor

	switch compressorName {
	case "Brotli":
		opts := cbrotli.WriterOptions{
			Quality: 11, // Compression level
			LGWin:   0,  // Automatic window
		}
		mc = ncd.NewManagedCompressorBrotli(opts)
	case "Gzip":
		mc = ncd.NewManagedCompressorGzip()
	case "zstd":
		opts := []zstd.EOption{
			zstd.WithEncoderLevel(zstd.SpeedBestCompression),
		}
		mc = ncd.NewManagedCompressorZstd(opts)
	}

	if *argZCmd != "" {
		zCmdArgs := strings.Fields(*argZCmd)
		compressorName = zCmdArgs[0] + " (external compressor)"

		mc = ncd.NewManagedCompressorExternal(zCmdArgs)
	}

	cx := ncd.CXVector(seqs, mc)
	cxx := ncd.CXXVector(seqs, mc)

	if *argStats {
		fmt.Println("COMPRESSOR")
		fmt.Println("==========")
		fmt.Println("Name:\t" + compressorName + "\n")
		fmt.Println("COMPRESSION METRICS")
		fmt.Println("===================")

		selfNCD := make([]float64, N)
		for i := range N {
			selfNCD[i] = ncd.NCD(cx[i], cx[i], cxx[i])
		}

		seqSize := make([]int, len(*seqs))
		for i, v := range *seqs {
			seqSize[i] = len(v)
		}

		writeStatsTable(os.Stdout, taxonNames, &seqSize, &cx, &selfNCD)
	}

	// Create the distance matrix
	D := ncd.NCDMatrix(seqs, &cx, mc)

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
