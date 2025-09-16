package ncd

import (
	"fmt"
	"io"
	"iter"
	"math"
)

/*
Holds the data of an "off-diagonal" triangular matrix of elements of type float64.

The elements of the diagonal are not included in the matrix.

Use NewODTriangularMatrix to create an instance.
*/
type TriangularMatrix struct {
	N       int
	RawData []float64 // Data is stored as a linear array representing the lower triangle
	Active  []bool    // Active series
}

func (m *TriangularMatrix) Copy() *TriangularMatrix {
	m2 := TriangularMatrix{m.N, make([]float64, m.N*(m.N-1)/2), make([]bool, m.N)}
	copy(m2.RawData, m.RawData)
	copy(m2.Active, m.Active)

	return &m2
}

/*
Creates a new off-diagonal triangular matrix

Parameters:

	n - dimension of the triangular matrix

Returns:

	Pointer to an instance of ODTriangularMatrix initialized with zero values
*/
func NewTriangularMatrix(n int) *TriangularMatrix {
	data := make([]float64, n*(n-1)/2)
	active := make([]bool, n)
	for i := range active {
		active[i] = true
	}

	return &TriangularMatrix{n, data, active}
}

/*
Returns the index of the underlying slice of the triangular matrix that corresponds to the off-diagonal position (i, j)
*/
func (m *TriangularMatrix) index(i int, j int) int {
	x := max(i, j)
	y := min(i, j)

	// if x > m.n {
	// 	nString := strconv.Itoa(m.n)
	// 	panic(
	// 		"tried to access position (" + strconv.Itoa(i) + ", " + strconv.Itoa(j) + ") in matrix of dimensions " +
	// 			nString + " x " + nString)
	// }

	return (x * (x - 1) / 2) + y
}

/*
Gets the value of the off-diagonal position (i, j) in the triangular matrix
*/
func (m *TriangularMatrix) Get(i int, j int) float64 {
	idx := m.index(i, j)

	return m.RawData[idx]
}

/*
Sets v as the value for position the off-diagonal position (i, j) in the triangular matrix
*/
func (m *TriangularMatrix) Set(i int, j int, v float64) {
	idx := m.index(i, j)
	m.RawData[idx] = v
}

func (m *TriangularMatrix) Sequence(i int) iter.Seq2[int, float64] {
	return func(yield func(int, float64) bool) {
		for j := range m.N {
			if m.Active[j] && (i != j) {
				if !yield(j, m.Get(i, j)) {
					return
				}
			}
		}
	}
}

// Returns the (row, column) index of the first instance of the smallest value in the matrix
func (m *TriangularMatrix) ArgMin() (int, int) {
	var min_i int
	var min_j int
	v_min := math.MaxFloat64

	for i := range m.N {
		for j, v := range m.Sequence(i) {
			if v < v_min {
				v_min = v
				min_i = i
				min_j = j
			}
		}
	}

	return min_i, min_j
}

func (m *TriangularMatrix) Show() {
	for i := range m.N {
		fmt.Printf("%d", i)
		if !m.Active[i] {
			fmt.Print("\n")
			continue
		}
		for j := range i {
			if i == j {
				continue
			}
			if !m.Active[j] {
				fmt.Print("\t")
			} else {
				fmt.Printf("\t%10.5f", m.Get(i, j))
			}
		}
		fmt.Print("\n")
	}

	fmt.Print("\t")
	for j := range m.N - 2 {
		fmt.Printf("%10d\t", j)
	}
	fmt.Printf("%10d\n", m.N-2)
}

func WriteLabelledTriangularMatrix(buf io.Writer, labels *[]string, M *TriangularMatrix, p int) (int, error) {
	b := 0              // Count of written bytes
	bb := 0             // Count of written bytes by a single write attempt
	fieldWidth := p + 2 // Field width for padding
	var s string
	if len(*labels) != M.N {
		return 0, fmt.Errorf("number of labels differs (%d) from the number of rows (%d)", len(*labels), M.N)
	}

	printWidth := 0
	for _, s := range *labels {
		l := len(s)
		if l > printWidth {
			printWidth = l
		}
	}

	for i := range M.N {
		s = fmt.Sprintf("%-*s", printWidth, (*labels)[i])
		bb, _ = fmt.Fprint(buf, s)
		b += bb
		for j := range i {
			if i == j {
				continue
			} else {
				s = fmt.Sprintf(" %-*.*g", fieldWidth, p, M.Get(i, j))
				bb, _ = fmt.Fprint(buf, s)
				b += bb
			}
		}
		bb, _ = fmt.Fprintf(buf, "\n")
		b += bb
	}

	return b, nil
}
