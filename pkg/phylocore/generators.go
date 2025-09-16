package phylocore

/*
Create a tree that is as balanced as possible for a given number of outer nodes (greater than 1).

Node IDs are assigned in PHYLIP order.

# Parameters
  - nbOuter: The number of outer nodes for the tree. Must be greater than 1.
*/
func MakeBalancedTree(taxa *TaxonSet) *Tree {
	tree := taxa.MakeUnassembledTree()
	nbOuter := taxa.Len()
	if nbOuter < 2 {
		panic("cannot make a balanced tree with less than 2 taxa")
	}
	bifurcate(tree, tree.Root, nbOuter, nbOuter+1, 0)

	return tree
}

func bifurcate(tree *Tree, node *Node, nbOuter int, nextIdInner int, nextIdOuter int) (int, int) {
	nbOuterRight := nbOuter / 2
	nbOuterLeft := nbOuter - nbOuterRight

	var id int

	if nbOuterLeft == 1 {
		id = nextIdOuter
		nextIdOuter += 1
	} else {
		id = nextIdInner
		nextIdInner += 1
	}
	leftChildNode := tree.Nodes[id]
	node.AddChild(leftChildNode, tree.NewBranch())

	if nbOuterLeft > 1 {
		nextIdInner, nextIdOuter = bifurcate(tree, leftChildNode, nbOuterLeft, nextIdInner, nextIdOuter)
	}

	if nbOuterRight == 1 {
		id = nextIdOuter
		nextIdOuter += 1
	} else {
		id = nextIdInner
		nextIdInner += 1
	}
	rightChildNode := tree.Nodes[id]
	node.AddChild(rightChildNode, tree.NewBranch())

	if nbOuterRight > 1 {
		nextIdInner, nextIdOuter = bifurcate(tree, rightChildNode, nbOuterRight, nextIdInner, nextIdOuter)
	}

	return nextIdInner, nextIdOuter
}

/*
Create a star tree, i.e. a tree with only one inner node (the root).

Node IDs are assigned in PHYLIP order.

# Parameters
  - nbOuter: The number of outer nodes for the tree. Must be greater than 1.
*/
func MakeStarTree(taxa *TaxonSet) *Tree {
	tree := taxa.MakeUnassembledTree()
	nbOuter := taxa.Len()
	if nbOuter < 2 {
		panic("cannot make a star tree with less than 2 taxa")
	}

	for i := range nbOuter {
		outerNode := tree.Nodes[i]
		tree.Root.AddChild(outerNode, tree.NewBranch())
	}

	return tree
}
