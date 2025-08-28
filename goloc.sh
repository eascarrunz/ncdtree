#!/usr/bin/env sh

# Get the number of lines of code (excluding blank lines and commented lines) in Go source files
# Usage: sh goloc.sh <PATH>
# If the path is a directory, the script counts all the lines of code in all the files in it with the ".go" extensions.

if [ -d $1 ]; then
    cat $(find $1 -name *.go) | \
    grep -v "^\s*$" | \
    grep -v "^//" | \
    sed -E ':a; /\/\*/{ :b; N; /\*\//!bb; s@/\*[^*]*\*+([^/*][^*]*\*+)*\/@@; ba }' | wc -l
else
    grep -v "^\s*$" $1 | \
    grep -v "^//" | \
    sed -E ':a; /\/\*/{ :b; N; /\*\//!bb; s@/\*[^*]*\*+([^/*][^*]*\*+)*\/@@; ba }' | wc -l
fi
