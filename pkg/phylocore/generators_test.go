package phylocore

import (
	"testing"
)

func makeTaxonSet(n int) *TaxonSet {
	names := make([]string, n)
	for i := range n {
		names[i] = string(rune('A' + i))
	}
	ts, _ := NewTaxonSet(names)
	return ts
}

func checkApproximateNodeBalance(node *Node) (bool, int) {
	/*
		Approximate node balance means that the node is as balanced as possible while allowing for trees with numbers of outer nodes that are not powers of 2.
		Approximate balance is fulfilled when the difference in the left and right number of descendants is 0 or 2
	*/

	outDegree := node.OutDegree()
	if outDegree == 0 {
		return true, 0
	}
	if outDegree != 2 {
		return false, outDegree
	}

	leftBalance, nbLeftDescendants := checkApproximateNodeBalance(node.Out[0].Child)
	rightBalance, nbRightDescendants := checkApproximateNodeBalance(node.Out[1].Child)

	isBalanced := false
	diffNbDescendants := nbLeftDescendants - nbRightDescendants

	if diffNbDescendants == 0 {
		isBalanced = true
	} else if diffNbDescendants > 0 {
		isBalanced = diffNbDescendants == 2
	} else {
		isBalanced = diffNbDescendants == -2
	}

	isBalanced = isBalanced && leftBalance && rightBalance

	return isBalanced, nbLeftDescendants + nbRightDescendants
}

// func checkNodeBalance(node *Node) bool {
// 	nbOuterChildren := 0
// 	nbInnerChildren := 0

// 	for i, branch := range node.Out {
// 		if i > 1 {
// 			return false
// 		}
// 		if branch.Child.IsInner() {
// 			nbInnerChildren += 1
// 		} else {
// 			nbOuterChildren += 1
// 		}
// 	}

// 	node.NbDescendants()
// }

func TestMakeBalancedTree(t *testing.T) {
	tests := []struct {
		nbOuter int
	}{
		{2},
		{3},
		{4},
		{5},
	}
	for _, tt := range tests {
		taxa := makeTaxonSet(tt.nbOuter)
		tree := MakeBalancedTree(taxa)

		isBalanced, _ := checkApproximateNodeBalance(tree.Root)

		if !isBalanced {
			t.Errorf("MakeBalancedTree(%d): tree is not balanced", tt.nbOuter)
		}
	}
}

func TestMakeStarTree(t *testing.T) {
	tests := []struct {
		nbOuter int
	}{
		{2},
		{5},
		{6},
	}
	for _, tt := range tests {
		taxa := makeTaxonSet(tt.nbOuter)
		tree := MakeStarTree(taxa)
		root := tree.Root
		if root == nil {
			t.Errorf("MakeStarTree(%d): root is nil", tt.nbOuter)
			continue
		}
		if root.OutDegree() != tt.nbOuter {
			t.Errorf("MakeStarTree(%d): root.OutDegree() = %d, want %d", tt.nbOuter, root.OutDegree(), tt.nbOuter)
		}
		for i := range tt.nbOuter {
			node := tree.Nodes[i]
			if node == root {
				continue
			}
			if node.In == nil {
				t.Errorf("MakeStarTree(%d): node %d has no incoming branch", tt.nbOuter, i)
			}
			if node.OutDegree() != 0 {
				t.Errorf("MakeStarTree(%d): node %d OutDegree() = %d, want 0", tt.nbOuter, i, node.OutDegree())
			}
		}
	}
}
