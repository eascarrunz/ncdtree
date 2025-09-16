package ncd

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

/*
Creates a vector with the compressed sizes of a list of sequences.
Vector element type is float64 for other NCD calculations.
*/
func CXVector(seqs *[][]byte, mc ManagedCompressor) []float64 {
	N := len(*seqs)
	cx := make([]float64, N)

	for i, s := range *seqs {
		mc.Send(s)
		cx[i] = float64(mc.Process())
	}

	return cx
}

/*
Creates a vector with the compressed sizes of a list of sequences concatenated with themselves.
Vector element type is float64 for consistance with CXVector.
*/
func CXXVector(seqs *[][]byte, mc ManagedCompressor) []float64 {
	N := len(*seqs)
	cxx := make([]float64, N)

	// mc.Process()

	for i, s := range *seqs {
		mc.Send(s)
		mc.Send(s)
		cxx[i] = float64(mc.Process())
	}

	return cxx
}

/*
Creates an NCD matrix from a list of sequences, using a pre-computed vector compressed sizes
*/
func NCDMatrix(seqs *[][]byte, cx *[]float64, mc ManagedCompressor) *TriangularMatrix {
	N := len(*seqs)
	D := NewTriangularMatrix(N)

	mc.Process()

	for i := 0; i < N; i += 1 {
		ca := (*cx)[i]
		for j := 0; j < i; j += 1 {
			cb := (*cx)[j]
			mc.Send((*seqs)[i])
			mc.Send((*seqs)[j])
			cab := float64(mc.Process())
			D.Set(i, j, NCD(ca, cb, cab))
		}
	}

	return D
}
