package phylocore

import (
	"math"
	"testing"
)

func TestPatristic(t *testing.T) {
	t.Run("BasicTree", func(t *testing.T) {
		// Construct: (A:1.0, (B:2.0, C:3.0):0.5);
		// Taxon IDs: A=0, B=1, C=2
		ts, _ := NewTaxonSet([]string{"A", "B", "C"})
		tree := NewEmptyTree(5)

		nodeA := tree.NewNode()
		nodeA.TaxonId = 0
		nodeB := tree.NewNode()
		nodeB.TaxonId = 1
		nodeC := tree.NewNode()
		nodeC.TaxonId = 2

		inner := tree.NewNode()
		root := tree.NewNode()
		tree.Root = root

		brA := tree.NewBranch()
		brA.Length = 1.0
		root.AddChild(nodeA, brA)

		brInner := tree.NewBranch()
		brInner.Length = 0.5
		root.AddChild(inner, brInner)

		brB := tree.NewBranch()
		brB.Length = 2.0
		inner.AddChild(nodeB, brB)

		brC := tree.NewBranch()
		brC.Length = 3.0
		inner.AddChild(nodeC, brC)

		d := ts.Patristic(tree)

		// Dist(B, C) = 2.0 + 3.0 = 5.0
		if got := d.Get(1, 2); got != 5.0 {
			t.Errorf("Dist(B, C) = %v, want 5.0", got)
		}
		// Dist(A, B) = 1.0 + 0.5 + 2.0 = 3.5
		if got := d.Get(0, 1); got != 3.5 {
			t.Errorf("Dist(A, B) = %v, want 3.5", got)
		}
		// Dist(A, C) = 1.0 + 0.5 + 3.0 = 4.5
		if got := d.Get(0, 2); got != 4.5 {
			t.Errorf("Dist(A, C) = %v, want 4.5", got)
		}
	})

	t.Run("MissingTaxon", func(t *testing.T) {
		// TaxonSet has A, B, C but tree only has A and B.
		ts, _ := NewTaxonSet([]string{"A", "B", "C"})
		tree := NewEmptyTree(3)
		a := tree.NewNode()
		a.TaxonId = 0
		b := tree.NewNode()
		b.TaxonId = 1
		root := tree.NewNode()
		tree.Root = root

		br1 := tree.NewBranch()
		br1.Length = 1.0
		root.AddChild(a, br1)
		br2 := tree.NewBranch()
		br2.Length = 1.0
		root.AddChild(b, br2)

		d := ts.Patristic(tree)
		// C is at index 2. Distances involving C should be NaN.
		if !math.IsNaN(d.Get(0, 2)) {
			t.Errorf("Dist(A, C) should be NaN, got %v", d.Get(0, 2))
		}
		if !math.IsNaN(d.Get(1, 2)) {
			t.Errorf("Dist(B, C) should be NaN, got %v", d.Get(1, 2))
		}
	})

	t.Run("InternalNodeTaxon", func(t *testing.T) {
		// Case where an internal node (root here) also carries a taxon identity.
		ts, _ := NewTaxonSet([]string{"Parent", "Child"})
		tree := NewEmptyTree(2)
		p := tree.NewNode()
		p.TaxonId = 0
		tree.Root = p

		c := tree.NewNode()
		c.TaxonId = 1

		br := tree.NewBranch()
		br.Length = 7.0
		p.AddChild(c, br)

		d := ts.Patristic(tree)
		if got := d.Get(0, 1); got != 7.0 {
			t.Errorf("Dist(Parent, Child) = %v, want 7.0", got)
		}
	})

	t.Run("StarTreeDistances", func(t *testing.T) {
		ts, _ := NewTaxonSet([]string{"A", "B", "C", "D"})
		tree := MakeStarTree(ts)
		for _, br := range tree.Branches {
			br.Length = 1.0
		}
		d := ts.Patristic(tree)
		// In a star tree with branch lengths 1.0, any pair distance is 2.0.
		if got := d.Get(0, 3); got != 2.0 {
			t.Errorf("Dist(A, D) = %v, want 2.0", got)
		}
	})

	t.Run("ZeroLength", func(t *testing.T) {
		ts, _ := NewTaxonSet([]string{"A", "B"})
		tree := MakeStarTree(ts)
		for _, br := range tree.Branches {
			br.Length = 0
		}
		d := ts.Patristic(tree)
		if got := d.Get(0, 1); got != 0 {
			t.Errorf("Dist(A, B) = %v, want 0", got)
		}
	})
}
