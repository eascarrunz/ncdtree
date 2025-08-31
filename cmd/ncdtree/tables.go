package main

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

func findStringMaxWidth(list *[]string) int {
	var w, l int
	for _, s := range *list {
		l = len(s)
		if l > w {
			w = l
		}
	}

	return w
}

func findIntMaxWidth(list *[]int) int {
	var w, l int
	for _, v := range *list {
		l = len(strconv.Itoa(v))
		if l > w {
			w = l
		}
	}

	return w
}

func fmtFloatField(x float64, precision int, maxWidth int) string {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return fmt.Sprintf("%.*g", precision, x)
	}

	abs := math.Abs(x)
	intDigits := 1
	if abs >= 1 {
		intDigits = int(math.Floor(math.Log10(abs))) + 1
	}
	frac := precision - intDigits
	if frac < 0 {
		frac = 0
	}

	plain := fmt.Sprintf("%.*f", frac, x) // do NOT trim zeros or dot
	if len(plain) <= maxWidth {
		return plain
	}

	// fall back to scientific; use %e with precision - 1 fraction digits so total sig = prec
	sci := fmt.Sprintf("%.*e", precision-1, x) // do NOT trim

	return sci
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func writeStatsTable(w io.Writer, taxonNames *[]string, seqLen *[]int, cx *[]float64, selfNCD *[]float64) {
	colTitles := [6]string{"#", "Taxon", "Size", "CompressedSize", "CompressionRatio", "SelfNCD"}
	colWidths := make(map[string]int, 6)
	n := len(*taxonNames)

	// Make sure that the columns will be at least as wide as their titles
	for _, s := range colTitles {
		colWidths[s] = len(s)
	}

	colWidths["#"] = len(strconv.Itoa(n))
	colWidths["Taxon"] = max(len("Taxon"), findStringMaxWidth(taxonNames))
	colWidths["Size"] = max(len("Size"), findIntMaxWidth(seqLen))
	colWidths["CompressedSize"] = max(len("CompressedSize"), colWidths["Size"])
	colWidths["CompressionRatio"] = max(colWidths["CompressionRatio"], 8)
	colWidths["SelfNCD"] = max(colWidths["SelfNCD"], 10)

	fieldGapSize := 2
	gapString := strings.Repeat(" ", fieldGapSize)

	tableWidth := len(colWidths)*fieldGapSize - fieldGapSize
	for _, v := range colWidths {
		tableWidth += v
	}

	// Print header
	isFirst := true
	for _, s := range colTitles {
		if isFirst {
			isFirst = false
		} else {
			fmt.Fprint(w, gapString)
		}
		fmt.Fprintf(w, "%-*s", colWidths[s], s)
	}

	fmt.Fprint(w, "\n")

	fmt.Fprintln(w, strings.Repeat("-", tableWidth))

	for i, taxon := range *taxonNames {
		fmt.Fprintf(w, "%-*.d", colWidths["#"], i+1)
		fmt.Fprint(w, gapString)
		fmt.Fprintf(w, "%-*s", colWidths["Taxon"]+fieldGapSize, taxon)
		fmt.Fprintf(w, "%-*d", colWidths["Size"]+fieldGapSize, (*seqLen)[i])
		fmt.Fprintf(w, "%-*g", colWidths["CompressedSize"]+fieldGapSize, (*cx)[i])
		// fmt.Fprintf(w, "%-.*g%s", colWidths["CompressionRatio"]-1, float64((*seqLen)[i])/(*cx)[i], gapString)
		s := padRight(fmtFloatField(float64((*seqLen)[i])/(*cx)[i], 8, colWidths["CompressionRatio"]), colWidths["CompressionRatio"])
		fmt.Fprintf(w, "%s%s", s, gapString)
		s = padRight(fmtFloatField((*selfNCD)[i], 6, colWidths["SelfNCD"]), colWidths["SelfNCD"])
		// fmt.Fprintf(w, "%-*.g", colWidths["SelfNCD"], (*selfNCD)[i])
		fmt.Fprintf(w, "%s%s", s, gapString)
		fmt.Fprintln(w)
	}
}
