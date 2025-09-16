package ncd

import (
	"reflect"
	"testing"
)

func TestNewTriangularMatrix(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{2, 2},
		{3, 3},
		{5, 5},
	}
	for _, tt := range tests {
		m := NewTriangularMatrix(tt.n)
		if m.N != tt.n {
			t.Errorf("NewTriangularMatrix(%d).N = %d, want %d", tt.n, m.N, tt.n)
		}
		if len(m.RawData) != tt.n*(tt.n-1)/2 {
			t.Errorf("NewTriangularMatrix(%d).RawData len = %d, want %d", tt.n, len(m.RawData), tt.n*(tt.n-1)/2)
		}
		for i := range m.Active {
			if !m.Active[i] {
				t.Errorf("NewTriangularMatrix(%d).Active[%d] = false, want true", tt.n, i)
			}
		}
	}
}

func TestGetSet(t *testing.T) {
	m := NewTriangularMatrix(4)
	tests := []struct {
		i, j int
		val  float64
	}{
		{1, 0, 1.1},
		{2, 0, 2.2},
		{2, 1, 3.3},
		{3, 0, 4.4},
		{3, 1, 5.5},
		{3, 2, 6.6},
	}
	for _, tt := range tests {
		m.Set(tt.i, tt.j, tt.val)
		got := m.Get(tt.i, tt.j)
		if got != tt.val {
			t.Errorf("Set/Get(%d,%d): got %v, want %v", tt.i, tt.j, got, tt.val)
		}
		// Test symmetry
		gotSym := m.Get(tt.j, tt.i)
		if gotSym != tt.val {
			t.Errorf("Get symmetry (%d,%d): got %v, want %v", tt.j, tt.i, gotSym, tt.val)
		}
	}
}

func TestCopy(t *testing.T) {
	m := NewTriangularMatrix(3)
	m.Set(2, 1, 7.7)
	m.Active[1] = false
	m2 := m.Copy()
	if !reflect.DeepEqual(m.RawData, m2.RawData) {
		t.Errorf("Copy: RawData not equal")
	}
	if !reflect.DeepEqual(m.Active, m2.Active) {
		t.Errorf("Copy: Active not equal")
	}
	if m2.Get(2, 1) != 7.7 {
		t.Errorf("Copy: value mismatch")
	}
}

func TestArgMin(t *testing.T) {
	m := NewTriangularMatrix(3)
	m.Set(1, 0, 5.5)
	m.Set(2, 0, 2.2)
	m.Set(2, 1, 3.3)
	i, j := m.ArgMin()
	if !(m.Get(i, j) <= m.Get(1, 0) && m.Get(i, j) <= m.Get(2, 1) && m.Get(i, j) <= m.Get(2, 0)) {
		t.Errorf("ArgMin: got (%d,%d)=%v, not minimum", i, j, m.Get(i, j))
	}
}
