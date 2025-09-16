package ncd

import (
	// "bytes"
	"testing"
)

type fakeCompressor struct {
	count int
}

func (fc *fakeCompressor) Send(data []byte) (int, error) {
	fc.count += len(data)
	return len(data), nil
}
func (fc *fakeCompressor) Process() int {
	out := fc.count
	fc.count = 0
	return out
}

func TestNCD_CXVector_CXXVector_NCDMatrix(t *testing.T) {
	seqInputs := [][]byte{
		[]byte("A"),
		[]byte("AA"),
		[]byte("AB"),
		[]byte("ABC"),
		[]byte("ABCD"),
		[]byte("AABBAABB"),
	}

	tests := []struct {
		name string
		seqs [][]byte
	}{
		{"single", seqInputs[:1]},
		{"two", seqInputs[:2]},
		{"three", seqInputs[:3]},
		{"four", seqInputs[:4]},
		{"five", seqInputs[:5]},
		{"six", seqInputs[:6]},
	}

	for _, tt := range tests {
		seqs := &tt.seqs
		mc := &fakeCompressor{}

		// Test CXVector
		cx := CXVector(seqs, mc)
		if len(cx) != len(tt.seqs) {
			t.Errorf("%s: CXVector len = %d, want %d", tt.name, len(cx), len(tt.seqs))
		}
		for i, s := range tt.seqs {
			if cx[i] != float64(len(s)) {
				t.Errorf("%s: CXVector[%d] = %v, want %v", tt.name, i, cx[i], float64(len(s)))
			}
		}

		// Test CXXVector
		mc = &fakeCompressor{}
		cxx := CXXVector(seqs, mc)
		if len(cxx) != len(tt.seqs) {
			t.Errorf("%s: CXXVector len = %d, want %d", tt.name, len(cxx), len(tt.seqs))
		}
		for i, s := range tt.seqs {
			if cxx[i] != float64(2*len(s)) {
				t.Errorf("%s: CXXVector[%d] = %v, want %v", tt.name, i, cxx[i], float64(2*len(s)))
			}
		}

		// Test NCDMatrix
		mc = &fakeCompressor{}
		D := NCDMatrix(seqs, &cx, mc)
		if D.N != len(tt.seqs) {
			t.Errorf("%s: NCDMatrix N = %d, want %d", tt.name, D.N, len(tt.seqs))
		}
		// Check symmetry and diagonal
		for i := 0; i < D.N; i++ {
			for j := 0; j < i; j++ {
				// For fake compressor, cab = len(seqs[i]) + len(seqs[j])
				cab := float64(len(tt.seqs[i]) + len(tt.seqs[j]))
				ca := float64(len(tt.seqs[i]))
				cb := float64(len(tt.seqs[j]))
				want := NCD(ca, cb, cab)
				got := D.Get(i, j)
				if got != want {
					t.Errorf("%s: NCDMatrix.Get(%d,%d) = %v, want %v", tt.name, i, j, got, want)
				}
			}
		}
	}
}

func TestNCD(t *testing.T) {
	tests := []struct {
		x, y, xy float64
		want     float64
	}{
		{1, 1, 2, 1},
		{2, 1, 2, 0.5},
		{1, 2, 2, 0.5},
		{2, 2, 3, 0.5},
		{3, 2, 4, 2.0 / 3.0},
		{2, 3, 4, 2.0 / 3.0},
	}
	for _, tt := range tests {
		got := NCD(tt.x, tt.y, tt.xy)
		if got != tt.want {
			t.Errorf("NCD(%v,%v,%v) = %v, want %v", tt.x, tt.y, tt.xy, got, tt.want)
		}
	}
}
