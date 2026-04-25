package phylocore

import (
	"math"
	"ncdtree/pkg/ncd"
	"sync"
)

// Return a matrix of patristic distances between taxa in the tree.
// Rows/columns are indexed by TaxonId. Entries for taxa absent from the tree are NaN.
func (taxset *TaxonSet) Patristic(tree *Tree) *ncd.TriangularMatrix {
	nbTaxa := taxset.Len()

	D := ncd.NewTriangularMatrix(nbTaxa)
	for i := range D.RawData {
		D.RawData[i] = math.NaN()
	}

	type distEntry struct {
		taxonId int
		dist    float64
	}

	// Pool of distEntry slices. Freed child slices are returned here and reused
	// for subsequent nodes, avoiding repeated allocation during traversal.
	pool := sync.Pool{New: func() any { return []distEntry(nil) }}

	// subtreeTaxa[node.Id] = slice of (taxonId, dist-to-node) for taxa in the subtree
	subtreeTaxa := make([][]distEntry, len(tree.Nodes))

	tree.TraverseNodes(func(node *Node) {
		// Pre-compute exact capacity so taxa never reallocates.
		capacity := 0
		for _, branch := range node.Out {
			capacity += len(subtreeTaxa[branch.Child.Id])
		}
		if node.TaxonId > -1 {
			capacity++
		}

		taxa := pool.Get().([]distEntry)
		if cap(taxa) < capacity {
			taxa = make([]distEntry, 0, capacity)
		} else {
			taxa = taxa[:0]
		}

		// Merge children's taxa, computing cross-subtree patristic distances.
		// childTaxa is the outer loop so D.Row(e2.taxonId) is precomputed once per
		// child taxon; taxa (accumulated) is the inner loop.
		for _, branch := range node.Out {
			child := branch.Child
			childTaxa := subtreeTaxa[child.Id]
			subtreeTaxa[child.Id] = nil

			for _, e2 := range childTaxa {
				rowE2 := D.Row(e2.taxonId)
				d2 := e2.dist + branch.Length
				for _, e1 := range taxa {
					rowE2.Set(e1.taxonId, e1.dist+d2)
				}
			}

			for _, e := range childTaxa {
				taxa = append(taxa, distEntry{e.taxonId, e.dist + branch.Length})
			}

			pool.Put(childTaxa[:0])
		}

		if node.TaxonId > -1 {
			rowNode := D.Row(node.TaxonId)
			for _, e := range taxa {
				rowNode.Set(e.taxonId, e.dist)
			}
			taxa = append(taxa, distEntry{node.TaxonId, 0.0})
		}

		subtreeTaxa[node.Id] = taxa
	}, PostOrder)

	return D
}
