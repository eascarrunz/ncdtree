package phylocore

import (
	"fmt"
	"math"
	"strings"
)

/*
Returns the Newick representation of the subtree rooted in the node, omitting the final semicolon.
*/
func (node *Node) _newick(b *strings.Builder) {
	if len(node.Out) > 1 {
		b.WriteString("(")
		isFirst := true
		for _, branch := range node.Out {
			if !isFirst {
				b.WriteString(",")
			}

			branch.Child._newick(b)

			if !math.IsNaN(branch.Length) {
				b.WriteString(":")
				fmt.Fprint(b, branch.Length)
			}
			isFirst = false
		}
		b.WriteString(")")
	}
	b.WriteString(node.Label)
}

/*
Returns the Newick representation of the subtree rooted in the node.

Reference for the Newick format: https://phylipweb.github.io/phylip/newicktree.html
*/
func (node *Node) NewickString() string {
	var b strings.Builder
	node._newick(&b)
	b.WriteString(";")

	return b.String()
}

/*
Returns the Newick representation of the tree.

References for the Newick format:
  - https://phylipweb.github.io/phylip/newicktree.html
  - https://en.wikipedia.org/wiki/Newick_format
*/
func (tree *Tree) NewickString() string {
	return tree.Root.NewickString()
}
