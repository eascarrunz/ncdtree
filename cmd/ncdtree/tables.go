package main

import (
	"fmt"
	"io"
	"math"
	"ncdtree/pkg/stats"
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

	compressionRatios := make([]float64, len(*seqLen))
	for i, l := range *seqLen {
		compressionRatios[i] = float64(l) / (*cx)[i]
	}

	// Make sure that the columns will be at least as wide as their titles
	for _, s := range colTitles {
		colWidths[s] = len(s)
	}

	colWidths["#"] = len(strconv.Itoa(n))
	colWidths["Taxon"] = max(len("Taxon"), findStringMaxWidth(taxonNames))
	colWidths["Size"] = max(len("Size"), findIntMaxWidth(seqLen)) + 3 // + 3 for float printing in the stats at the end
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

	fmt.Fprintln(w, strings.Repeat("—", tableWidth))

	// Taxon data
	for i, taxon := range *taxonNames {
		fmt.Fprintf(w, "%-*.d", colWidths["#"], i+1)
		fmt.Fprint(w, gapString)
		fmt.Fprintf(w, "%-*s", colWidths["Taxon"]+fieldGapSize, taxon)
		fmt.Fprintf(w, "%-*d", colWidths["Size"]+fieldGapSize, (*seqLen)[i])
		fmt.Fprintf(w, "%-*g", colWidths["CompressedSize"]+fieldGapSize, (*cx)[i])
		// fmt.Fprintf(w, "%-.*g%s", colWidths["CompressionRatio"]-1, float64((*seqLen)[i])/(*cx)[i], gapString)
		s := padRight(fmtFloatField(compressionRatios[i], 8, colWidths["CompressionRatio"]), colWidths["CompressionRatio"])
		fmt.Fprintf(w, "%s%s", s, gapString)
		s = padRight(fmtFloatField((*selfNCD)[i], 6, colWidths["SelfNCD"]), colWidths["SelfNCD"])
		// fmt.Fprintf(w, "%-*.g", colWidths["SelfNCD"], (*selfNCD)[i])
		fmt.Fprintf(w, "%s", s)
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, strings.Repeat("-", tableWidth))

	colWidths["StatTitle"] = colWidths["#"] + colWidths["Taxon"] + fieldGapSize

	// Summary stats, going case by case because of different column types
	var colStat float64 // Column statistic
	var s string

	// Mean
	fmt.Fprintf(w, "%*s%s", colWidths["StatTitle"], "Mean", gapString)
	colStat = stats.MeanInt(seqLen)
	s = padRight(fmtFloatField(colStat, colWidths["Size"]-3, colWidths["Size"]), colWidths["Size"])
	fmt.Fprintf(w, "%s%s", s, gapString)
	colStat = stats.MeanFloat64(cx)
	s = padRight(fmtFloatField(colStat, colWidths["CompressedSize"]-3, colWidths["CompressedSize"]), colWidths["CompressedSize"])
	fmt.Fprintf(w, "%s%s", s, gapString)
	colStat = stats.MeanFloat64(&compressionRatios)
	s = padRight(fmtFloatField(colStat, colWidths["CompressionRatio"]-3, colWidths["CompressionRatio"]), colWidths["CompressionRatio"])
	fmt.Fprintf(w, "%s%s", s, gapString)
	colStat = stats.MeanFloat64(selfNCD)
	s = padRight(fmtFloatField(colStat, colWidths["SelfNCD"]-3, colWidths["SelfNCD"]), colWidths["SelfNCD"])
	fmt.Fprintf(w, "%s", s)
	fmt.Fprintln(w)

	// Median
	fmt.Fprintf(w, "%*s%s", colWidths["StatTitle"], "Median", gapString)
	colStat = stats.Median(seqLen)
	fmt.Fprintf(w, "%-*g%s", colWidths["Size"], colStat, gapString)
	colStat = stats.Median(cx)
	fmt.Fprintf(w, "%-*g%s", colWidths["CompressedSize"], colStat, gapString)
	colStat = stats.Median(&compressionRatios)
	s = padRight(fmtFloatField(colStat, 8, colWidths["CompressionRatio"]), colWidths["CompressionRatio"])
	fmt.Fprintf(w, "%s%s", s, gapString)
	colStat = stats.Median(selfNCD)
	s = padRight(fmtFloatField(colStat, 6, colWidths["SelfNCD"]), colWidths["SelfNCD"])
	fmt.Fprintf(w, "%s", s)
	fmt.Fprintln(w)

	// Minimum
	fmt.Fprintf(w, "%*s%s", colWidths["StatTitle"], "Minimum", gapString)
	colStat = float64(stats.Minimum(seqLen))
	fmt.Fprintf(w, "%-*g%s", colWidths["Size"], colStat, gapString)
	colStat = stats.Minimum(cx)
	fmt.Fprintf(w, "%-*g%s", colWidths["CompressedSize"], colStat, gapString)
	colStat = stats.Minimum(&compressionRatios)
	s = padRight(fmtFloatField(colStat, 8, colWidths["CompressionRatio"]), colWidths["CompressionRatio"])
	fmt.Fprintf(w, "%s%s", s, gapString)
	colStat = stats.Minimum(selfNCD)
	s = padRight(fmtFloatField(colStat, 6, colWidths["SelfNCD"]), colWidths["SelfNCD"])
	fmt.Fprintf(w, "%s", s)
	fmt.Fprintln(w)

	// Maximum
	fmt.Fprintf(w, "%*s%s", colWidths["StatTitle"], "Maximum", gapString)
	colStat = float64(stats.Maximum(seqLen))
	fmt.Fprintf(w, "%-*g%s", colWidths["Size"], colStat, gapString)
	colStat = stats.Maximum(cx)
	fmt.Fprintf(w, "%-*g%s", colWidths["CompressedSize"], colStat, gapString)
	colStat = stats.Maximum(&compressionRatios)
	s = padRight(fmtFloatField(colStat, 8, colWidths["CompressionRatio"]), colWidths["CompressionRatio"])
	fmt.Fprintf(w, "%s%s", s, gapString)
	colStat = stats.Maximum(selfNCD)
	s = padRight(fmtFloatField(colStat, 6, colWidths["SelfNCD"]), colWidths["SelfNCD"])
	fmt.Fprintf(w, "%s", s)
	fmt.Fprintln(w)

}
