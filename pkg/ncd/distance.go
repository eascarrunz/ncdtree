package ncd

func NCD(x float64, y float64, xy float64) float64 {
	if x > y {
		return (xy - y) / x
	} else {
		return (xy - x) / y
	}
}

func CXVector(seqs *[][]byte, mc ManagedCompressor) []float64 {
	N := len(*seqs)
	cx := make([]float64, N)

	// mc.Process()

	for i, s := range *seqs {
		mc.Send(s)
		cx[i] = float64(mc.Process())
	}

	return cx
}

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
