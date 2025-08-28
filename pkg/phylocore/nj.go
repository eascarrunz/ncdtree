package phylocore

import (
	"math"
	"ncdtree/pkg/ncd"
)

func updateR(D *ncd.TriangularMatrix, R *[]float64) {
	for i, v := range D.Active {
		(*R)[i] = 0.0
		if v {
			for j := 0; j < i; j += 1 {
				if !D.Active[j] {
					continue
				}
				d := D.Get(i, j)
				(*R)[i] += d
				(*R)[j] += d
			}
		}
	}
}

func updateD(D *ncd.TriangularMatrix, a int, b int, d_ca float64, d_cb float64) {
	for k := range D.N {
		if !D.Active[k] || k == a || k == b {
			continue
		}

		d_ak := D.Get(a, k)
		d_bk := D.Get(b, k)

		// Yang 2014, eq. 3.10
		d_ck := 0.5 * (d_ak + d_bk - d_ca - d_cb)
		D.Set(a, k, d_ck)
	}
}

// Select the indices that minimize Q
func selectJoinTargets(D *ncd.TriangularMatrix, R *[]float64, m float64) (int, int, float64) {
	Q_min := math.MaxFloat64
	a := -1
	b := -1
	d_ab := 0.0

	for i := range D.N {
		if !D.Active[i] {
			continue
		}
		for j := 0; j < i; j += 1 {
			if !D.Active[j] {
				continue
			}

			d_ij := D.Get(i, j)

			// Yang 2014, eq. 3.8
			q := (m-2)*d_ij - (*R)[i] - (*R)[j]

			if q < Q_min {
				Q_min = q
				a = i
				b = j
				d_ab = d_ij
			}
		}
	}

	return a, b, d_ab
}

func selectLastTargets(D *ncd.TriangularMatrix) (int, int, float64) {
	a := -1
	b := -1
	d_ab := 0.0

	for i := range D.N {
		if !D.Active[i] {
			continue
		}
		a = i
		for j := (D.N - 1); j > i; j -= 1 {
			if !D.Active[j] {
				continue
			}

			b = j
			d_ab = D.Get(i, j)
			break
		}
		break
	}

	return a, b, d_ab
}

func NeighbourJoining(taxset *TaxonSet, D *ncd.TriangularMatrix) *Tree {
	nbTaxa := taxset.Len()
	nbNode := 2*nbTaxa - 2
	tree := MakeUnassembledTree(nbNode)
	tree.Root = tree.Nodes[nbTaxa]

	// List that matches up nodes in the tree to the positions in the D matrix
	targetNodes := make([]*Node, nbTaxa)

	activeIndices := make([]bool, nbTaxa)
	var node_c *Node

	for i := range nbTaxa {
		node := tree.Nodes[i]
		node.TaxonId = i
		node.Label = taxset.GetName(i)
		targetNodes[i] = node
		activeIndices[i] = true
	}

	// Vector of R_x values
	R := make([]float64, nbTaxa)

	m := float64(nbTaxa)

	// c is the index of the inner node chosen to make the join at each iteration
	// The reverse order is so that the last node added is the root
	for c := (nbNode - 1); c >= nbTaxa; c -= 1 {
		updateR(D, &R)
		a, b, d_ab := selectJoinTargets(D, &R, m)

		node_a := targetNodes[a]
		node_b := targetNodes[b]
		node_c = tree.Nodes[c]
		branch_ca := tree.NewBranch()
		branch_cb := tree.NewBranch()

		// Yang 2014, eq. 3.9
		d_ca := 0.5 * (d_ab + (R[a]-R[b])/(m-2.0))

		d_cb := d_ab - d_ca
		branch_ca.Length = d_ca
		branch_cb.Length = d_cb

		node_c.AddChild(node_a, branch_ca)
		node_c.AddChild(node_b, branch_cb)

		// Update number of active rows in the matrix
		updateD(D, a, b, d_ca, d_cb)
		D.Active[b] = false

		// Update the list of targets
		targetNodes[a] = node_c
		targetNodes[b] = nil

		m -= 1.0
	}

	a, b, d_ab := selectLastTargets(D)
	node_a := targetNodes[a]
	node_b := targetNodes[b]
	branch_ab := tree.NewBranch()
	branch_ab.Length = d_ab

	switch tree.Root {
	case node_a:
		node_a.AddChild(node_b, branch_ab)
	case node_b:
		node_b.AddChild(node_a, branch_ab)
	default:
		panic("the root was not used in the last join")
	}

	return tree
}
