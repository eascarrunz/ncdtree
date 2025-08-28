# ncdtree

This is a demo of phylogenetic inference without alignment using the normalized compression distance (NCD) and neighbour-joining.

> [!WARNING]
> This project is for demonstration purposes.

<center>🚧 Work in progress 🚧</center>

## Normalized Compression Distance

Cilibrasi & Vitányi (2005) defined the normalized compression distance ($NCD_Z$) of strings $x$ and $y$ with a compressor $Z$ as:

$$
{NCD}_Z(x, y) = \dfrac{C_Z(xy) - min\{ C_Z(x), C_Z(y) \}}{max\{ C_Z(x), C_Z(y) \}}
$$

Where $C_Z(\circ)$ is the compressed size of a string with a compressor $Z$, and $xy$ is the concatenation of $x$ and $y$.

## Usage

Get a distance matrix and neighbour-joining tree in Newick format printed to `stdout`.

```sh
./ncdtree <SEQUENCES>
```

The \<SEQUENCES\> file must contain nucleotide or amino-acid sequences in Fasta format.

Get a neighbour joining tree in Newick format printed to `stdout`.

```sh
./nj <MATRIX>
```

The \<SEQUENCES\> file must contain a distance matrix in plaintext format:

```
taxon_a 	0 	5 	9 	9 	8
taxon_b 	5 	0 	10 	10 	9
taxon_c 	9 	10 	0 	8 	7
taxon_d 	9 	10 	8 	0 	3
taxon_e 	8 	9 	7 	3 	0
```

There must be no header, and the first column must contain the taxon names. The fields are separated by whitespace. Only the lower triangle of the matrix is read. The diagonal and the upper triangle of the matrix can be omitted.

## Build

Coming soon.
