package stats

import (
	"math"
	"slices"
)

type Numeric interface {
	~int | ~float64
}

// Computes the variance of a slice of numeric values (int or float64)
func Variance[T Numeric](X *[]T, sample bool) float64 {
	if len(*X) < 2 {
		return 0.0
	}

	K := float64((*X)[0]) // Shift constant
	var Ex, Ex2 float64

	for _, x := range *X {
		xfloat := float64(x)
		Ex += xfloat - K
		Ex2 += math.Pow(xfloat-K, 2)
	}

	n := float64(len(*X))
	denom := n
	if sample {
		denom -= 1.0
	}

	return (Ex2 - math.Pow(Ex, 2)/n) / denom
}

// Computes the sum of a slice of float64 values using Kahan's compensation algorithm
func KahanSum(X *[]float64) float64 {
	sum := 0.0
	c := 0.0 // Compensation

	for _, x := range *X {
		y := x - c
		t := sum + y
		c = (t - sum) - y
		sum = t
	}

	return sum
}

// Computes the mean of a slice of float64 values,
func MeanFloat64(X *[]float64) float64 {
	return KahanSum(X) / float64(len(*X))
}

// Computes the mean of a slice of int values
func MeanInt(X *[]int) float64 {
	xSum := 0
	for _, v := range *X {
		xSum += v
	}

	return float64(xSum) / float64(len(*X))
}

// Computes the median of a slice of numeric values (int or float64)
func Median[T Numeric](X *[]T) float64 {
	tmp := slices.Clone(*X)
	slices.Sort(tmp)

	n := len(tmp)
	if n%2 == 0 {
		return (float64(tmp[n/2]) + float64(tmp[(n/2)+1])) / 2.0
	} else {
		return float64(tmp[(n+1)/2])
	}
}

// Returns the minimum value in a slice of int or float values
func Minimum[T Numeric](X *[]T) T {
	N := len(*X)
	if N == 1 {
		return (*X)[0]
	}
	minX := (*X)[0]

	for i := 1; i < N; i += 1 {
		if minX > (*X)[i] {
			minX = (*X)[i]
		}
	}

	return minX
}

// Returns the maximum value in a slice of int or float values
func Maximum[T Numeric](X *[]T) T {
	N := len(*X)
	if N == 1 {
		return (*X)[0]
	}
	maxX := (*X)[0]

	for i := 1; i < N; i += 1 {
		if maxX < (*X)[i] {
			maxX = (*X)[i]
		}
	}

	return maxX
}
