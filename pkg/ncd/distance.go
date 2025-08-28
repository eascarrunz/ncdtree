package ncd

func NCD(x float64, y float64, xy float64) float64 {
	if x > y {
		return (xy - y) / x
	} else {
		return (xy - x) / y
	}
}

func CXVector(seqs *[][]byte, ctx CompressionContext) []float64 {
	N := len(*seqs)
	cx := make([]float64, N)

	ctx.SizeReset()

	for i, s := range *seqs {
		ctx.Write(s)
		cx[i] = float64(ctx.SizeReset())
	}

	return cx
}

func CXXVector(seqs *[][]byte, ctx CompressionContext) []float64 {
	N := len(*seqs)
	cxx := make([]float64, N)

	ctx.SizeReset()

	for i, s := range *seqs {
		ctx.Write(s)
		ctx.Write(s)
		cxx[i] = float64(ctx.SizeReset())
	}

	return cxx
}

func NCDMatrix(seqs *[][]byte, cx *[]float64, ctx *CompressionContext) *TriangularMatrix {
	N := len(*seqs)
	D := NewTriangularMatrix(N)

	ctx.SizeReset()

	for i := 0; i < N; i += 1 {
		ca := (*cx)[i]
		for j := 0; j < i; j += 1 {
			cb := (*cx)[j]
			ctx.Write((*seqs)[i])
			ctx.Write((*seqs)[j])
			cab := float64(ctx.SizeReset())
			D.Set(i, j, NCD(ca, cb, cab))
		}
	}

	return D
}
