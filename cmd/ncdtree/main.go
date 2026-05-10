package main

import (
	"bufio"
	"fmt"
	"math/bits"
	"ncdtree/pkg/fasta"
	"ncdtree/pkg/ncd"
	"ncdtree/pkg/phylocore"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/akamensky/argparse"
)

func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	// bits.Len(uint(n-1)) gives the exponent needed.
	return 1 << bits.Len(uint(n-1))
}

const inputBufSize = 64 * 1024

func setupThreads(t int) {
	tMax := runtime.NumCPU()

	if t > tMax {
		fmt.Fprintf(os.Stderr, "Requested %d threads, but only %d threads are available", t, tMax)
	}

	runtime.GOMAXPROCS(t)
}

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
		&argparse.Options{Required: false, Default: "", Help: "External compressor command (overrides -Z). Must read from stdin and write compressed data to stdout, e.g. 'gzip -c'"},
	)
	argStats := parser.Flag(
		"s", "stats",
		&argparse.Options{Required: false, Help: "Print statistics"},
	)
	argThreads := parser.Int(
		"t", "threads",
		&argparse.Options{
			Required: false,
			Validate: func(args []string) error {
				t, err := strconv.Atoi(args[0])
				if err != nil {
					return err
				}
				if t < 1 {
					return fmt.Errorf("Invalid threads value: %d. Must be greater than 1", t)
				}
				return nil
			},
			Default: runtime.GOMAXPROCS(-1),
			Help:    "Number of threads to use for parallel compression. The default value is auto-selected by the Go runtime, but might not be optimal for all workloads"},
	)
	argNoTree := parser.Flag(
		"", "notree",
		&argparse.Options{Required: false, Help: "Do not estimate a tree. Only write out distance matrix."},
	)

	parser.Parse(os.Args)

	var input *os.File
	var err error
	var taxonNames []string
	var seqs [][]byte
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

	setupThreads(*argThreads)

	N := len(taxonNames)

	compressorName := *argAlgo

	var mcFactory ncd.ManagedCompressorFactory

	maxSeqLength := 0

	for _, seq := range seqs {
		l := len(seq)
		if l > maxSeqLength {
			maxSeqLength = l
		}
	}

	targetWindowSize := nextPowerOfTwo(maxSeqLength * 2)
	windowSize, err := ncd.ComputeCompressorWindowSize(compressorName, targetWindowSize)
	if err != nil {
		panic(err)
	}

	switch compressorName {
	case "Brotli":
		mcFactory = ncd.NewManagedCompressorBrotliFactory(windowSize)
	case "Gzip":
		mcFactory = ncd.NewManagedCompressorGzipFactory()
	case "zstd":
		mcFactory = ncd.NewManagedCompressorZstdFactory(windowSize)
	}

	if *argZCmd != "" {
		zCmdArgs := strings.Fields(*argZCmd)
		compressorName = zCmdArgs[0] + " (external compressor)"

		mcFactory = ncd.NewManagedCompressorExternalFactory(zCmdArgs)
	}

	cx := ncd.CXVector(seqs, mcFactory)
	cxx := ncd.CXXVector(seqs, mcFactory)

	if *argStats {
		fmt.Println("COMPRESSOR")
		fmt.Println("==========")
		fmt.Println("Name:\t" + compressorName)
		fmt.Printf("Window Size:\t%d\n", windowSize)
		fmt.Printf("Required Window Size:\t%d\n\n", targetWindowSize)
		fmt.Println("COMPRESSION METRICS")
		fmt.Println("===================")

		selfNCD := make([]float64, N)
		for i := range N {
			selfNCD[i] = ncd.NCD(cx[i], cx[i], cxx[i])
		}

		seqSize := make([]int, len(seqs))
		for i, v := range seqs {
			seqSize[i] = len(v)
		}

		writeStatsTable(os.Stdout, taxonNames, seqSize, cx, selfNCD)
	}

	// Create the distance matrix
	D := ncd.NCDMatrix(seqs, cx, mcFactory)

	outFileMatrix, err := os.Create("ncd_matrix.txt")
	if err != nil {
		panic(err)
	}
	defer outFileMatrix.Close()
	ncd.WriteLabelledTriangularMatrix(outFileMatrix, taxonNames, D, 9)

	if !*argNoTree {
		taxset, err := phylocore.NewTaxonSet(taxonNames)
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
