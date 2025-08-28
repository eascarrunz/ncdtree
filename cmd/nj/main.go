package main

import (
	"bufio"
	"fmt"
	"ncdtree/pkg/phylocore"
	"os"
)

func main() {
	var input *os.File
	var err error

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

	taxa, d := phylocore.ReadDistanceMatrix(*scanner)

	tree := phylocore.NeighbourJoining(taxa, d)

	fmt.Print(tree.NewickString(), "\n")
}
