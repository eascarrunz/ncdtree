package phylocore

import (
	"fmt"
	"math"
	"strings"
)

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
Return the Newick representation of the subtree rooted in the node.
*/
func (node *Node) NewickString() string {
	var b strings.Builder
	node._newick(&b)
	b.WriteString(";")

	return b.String()
}

/*
Return the Newick representation of the tree.
*/
func (tree *Tree) NewickString() string {
	return tree.Root.NewickString()
}
