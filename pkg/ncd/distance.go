package ncd

import (
	"runtime"
	"sync"
)

/*
Compute the NCD distance between two sequences of symbols based on their compressed sizes x and y, and their joint (concatenated) compressed size xy.

Formula after Cilibrasi & Vitányi (2005): NCD(x, y, xy) = (xy - min(x, y)) / max(x, y)
*/
func NCD(x float64, y float64, xy float64) float64 {
	if x > y {
		return (xy - y) / x
	} else {
		return (xy - x) / y
	}
}

func fillCVector(v []float64, f func(i int, mc ManagedCompressor) float64, factory ManagedCompressorFactory) {
	N := len(v)

	var wg sync.WaitGroup
	nWorkers := runtime.GOMAXPROCS(-1)
	jobsChannel := make(chan int, nWorkers)

	for w := 0; w < nWorkers; w += 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			mc := factory()

			for i := range jobsChannel {
				v[i] = f(i, mc)
			}
		}()
	}

	for i := range N {
		jobsChannel <- i
	}

	close(jobsChannel)

	wg.Wait()
}

/*
Creates a vector with the compressed sizes of a list of sequences.
Vector element type is float64 for other NCD calculations.
*/
func CXVector(seqs [][]byte, factory ManagedCompressorFactory) []float64 {
	N := len(seqs)
	cx := make([]float64, N)

	f := func(i int, mc ManagedCompressor) float64 {
		s := seqs[i]
		mc.Send(s)

		return float64(mc.Process())
	}

	fillCVector(cx, f, factory)

	return cx
}

/*
Creates a vector with the compressed sizes of a list of sequences concatenated with themselves.
Vector element type is float64 for consistance with CXVector.
*/
func CXXVector(seqs [][]byte, factory ManagedCompressorFactory) []float64 {
	N := len(seqs)
	cxx := make([]float64, N)

	f := func(i int, mc ManagedCompressor) float64 {
		s := seqs[i]
		mc.Send(s)
		mc.Send(s)

		return float64(mc.Process())
	}

	fillCVector(cxx, f, factory)

	return cxx
}

/*
Creates an NCD matrix from a list of sequences, using a pre-computed vector compressed sizes
*/
func NCDMatrix(seqs [][]byte, cx []float64, factory ManagedCompressorFactory) *TriangularMatrix {
	N := len(seqs)
	D := NewTriangularMatrix(N)

	var wg sync.WaitGroup
	nWorkers := runtime.GOMAXPROCS(-1)
	jobsChannel := make(chan int, nWorkers)

	for w := 0; w < nWorkers; w += 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mc := factory()

			for i := range jobsChannel {
				ca := cx[i]
				for j := 0; j < i; j += 1 {
					mc.Send(seqs[i])
					mc.Send(seqs[j])
					cab := float64(mc.Process())
					D.Set(i, j, NCD(ca, cx[j], cab))
				}
			}
		}()
	}

	for i := range N {
		jobsChannel <- i
	}
	close(jobsChannel)

	wg.Wait()

	return D
}
