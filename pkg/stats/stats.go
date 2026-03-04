package stats

import (
	"fmt"
	"math"
	"slices"
)

type Numeric interface {
	~int | ~float64
}

// Computes the variance of a slice of numeric values (int or float64)
func Variance[T Numeric](X *[]T, sample bool) float64 {
	n := len(*X)
	if n < 2 {
		return 0.0
	}

	var mean, M2 float64
	for i, x := range *X {
		delta := float64(x) - mean
		mean += delta / float64(i+1)
		M2 += delta * (float64(x) - mean)
	}

	denom := float64(n)

	// Bessel's correction
	if sample {
		denom -= 1.0
	}
	return M2 / denom
}

// Computes the standard deviation of numeric values (int or float64)
// The simple biased correction is used for the sampled variant
func StandardDeviation[T Numeric](X *[]T, sample bool) float64 {
	return math.Sqrt(Variance(X, sample))
}

func Covariance[T Numeric](X *[]T, Y *[]T, sample bool) float64 {
	if len(*X) != len(*Y) {
		panic(fmt.Sprintf("cannot compute covariance between vectors of different lengths: %d and %d", len(*X), len(*Y)))
	}
	n := len(*X)
	if n < 2 {
		return 0
	}

	var meanX, meanY, C float64
	for i := range *X {
		x := float64((*X)[i])
		y := float64((*Y)[i])
		dx := x - meanX
		meanX += dx / float64(i+1)
		meanY += (y - meanY) / float64(i+1)
		C += dx * (y - meanY)
	}

	denom := float64(n)

	// Bessel's correction
	if sample {
		denom -= 1.0
	}
	return C / denom
}

// Return the Pearson correlation coefficient between two variables
func CorrPearson[T Numeric](X *[]T, Y *[]T, sample bool) float64 {
	rho := Covariance(X, Y, sample)
	rho /= StandardDeviation(X, sample) * StandardDeviation(Y, sample)

	rho = min(rho, 1.0)
	rho = max(rho, -1.0)

	return rho
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
	if len((*X)) == 0 {
		panic("cannot compute median of an empty array")
	}
	tmp := slices.Clone(*X)
	slices.Sort(tmp)

	n := len(tmp)
	if n%2 == 0 {
		return (float64(tmp[(n/2)-1]) + float64(tmp[n/2])) / 2.0
	} else {
		return float64(tmp[n/2])
	}
}

// Returns the minimum value in a slice of int or float values
func Minimum[T Numeric](X *[]T) T {
	N := len(*X)
	if N == 0 {
		panic("cannot compute minimum of an empty array")
	}
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
	if N == 0 {
		panic("cannot compute maximum of an empty array")
	}
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
