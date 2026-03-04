package phylocore

import (
	"bufio"
	"math"
	"strings"
	"testing"
)

// readNewickString parses a Newick string and returns the tree and taxon set.
func readNewickString(s string) (*Tree, *TaxonSet, error) {
	r := bufio.NewReader(strings.NewReader(s))
	return ReadNewick(r)
}

// readNewickStringWithTaxset parses a Newick string against a known taxon set.
func readNewickStringWithTaxset(s string, taxset *TaxonSet, addNew bool) (*Tree, error) {
	r := bufio.NewReader(strings.NewReader(s))
	return taxset.ReadNewick(r, addNew)
}

// nodesByLabel builds a label → *Node map for a tree.
func nodesByLabel(tree *Tree) map[string]*Node {
	m := make(map[string]*Node, len(tree.Nodes))
	for _, n := range tree.Nodes {
		if n.Label != "" {
			m[n.Label] = n
		}
	}
	return m
}

func TestTreeNewickString(t *testing.T) {
    tests := []struct {
        name string
        setup func() *Tree
        want string
    }{
        {
            name: "single node",
            setup: func() *Tree {
                n := &Node{Label: "A"}
                return &Tree{Root: n}
            },
            want: "A;",
        },
        {
            name: "star tree",
            setup: func() *Tree {
                root := &Node{Label: ""}
                a := &Node{Label: "A"}
                b := &Node{Label: "B"}
                c := &Node{Label: "C"}
                root.Out = []*Branch{
                    {Child: a, Length: math.NaN()},
                    {Child: b, Length: math.NaN()},
                    {Child: c, Length: math.NaN()},
                }
                return &Tree{Root: root}
            },
            want: "(A,B,C);",
        },
        {
            name: "binary tree with branch lengths",
            setup: func() *Tree {
                root := &Node{Label: ""}
                a := &Node{Label: "A"}
                b := &Node{Label: "B"}
                c := &Node{Label: "C"}
                // (A:1.2,B:2.3):3.4,C:4.5
                left := &Node{Label: ""}
                left.Out = []*Branch{
                    {Child: a, Length: 1.2},
                    {Child: b, Length: 2.3},
                }
                root.Out = []*Branch{
                    {Child: left, Length: 3.4},
                    {Child: c, Length: 4.5},
                }
                return &Tree{Root: root}
            },
            want: "((A:1.2,B:2.3):3.4,C:4.5);",
        },
    }

	for _, tt := range tests {
		tree := tt.setup()
		got := tree.NewickString()
		if got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestReadNewick_RoundTrip parses a Newick string and verifies that serializing
// it back produces the same string.
func TestReadNewick_RoundTrip(t *testing.T) {
	cases := []string{
		"A;",
		"(A,B);",
		"((A,B),C);",
		"(A,B,C);",
		"(A:1,B:2);",
		"((A:1,B:2):3,C:4);",
		"(A,B)root;",
	}
	for _, input := range cases {
		tree, _, err := readNewickString(input)
		if err != nil {
			t.Errorf("%q: unexpected error: %v", input, err)
			continue
		}
		got := tree.NewickString()
		if got != input {
			t.Errorf("%q: round-trip got %q", input, got)
		}
	}
}

// TestReadNewick_TaxonIds checks that outer-node taxa are registered in the
// taxon set and that TaxonId fields are set correctly.
func TestReadNewick_TaxonIds(t *testing.T) {
	tree, taxset, err := readNewickString("(A,B,C);")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if taxset.Len() != 3 {
		t.Fatalf("taxset.Len() = %d, want 3", taxset.Len())
	}
	nodes := nodesByLabel(tree)
	for _, label := range []string{"A", "B", "C"} {
		n, ok := nodes[label]
		if !ok {
			t.Errorf("node %q not found", label)
			continue
		}
		id, exists := taxset.GetId(label)
		if !exists {
			t.Errorf("taxon %q not in taxset", label)
			continue
		}
		if n.TaxonId != id {
			t.Errorf("node %q: TaxonId = %d, want %d", label, n.TaxonId, id)
		}
	}
}

// TestReadNewick_InnerNodeLabel checks that a labelled inner node is NOT added
// to the taxon set in OuterNodes mode (the default).
func TestReadNewick_InnerNodeLabel(t *testing.T) {
	tree, taxset, err := readNewickString("(A,B)inner;")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if taxset.Len() != 2 {
		t.Errorf("taxset.Len() = %d, want 2 (inner label must not be added)", taxset.Len())
	}
	nodes := nodesByLabel(tree)
	inner, ok := nodes["inner"]
	if !ok {
		t.Fatal("inner node not found")
	}
	if inner.TaxonId != -1 {
		t.Errorf("inner.TaxonId = %d, want -1", inner.TaxonId)
	}
}

// TestReadNewick_BranchLengths verifies that branch lengths are parsed correctly.
func TestReadNewick_BranchLengths(t *testing.T) {
	tree, _, err := readNewickString("((A:1,B:2):3,C:4);")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodes := nodesByLabel(tree)
	cases := []struct {
		label string
		want  float64
	}{
		{"A", 1},
		{"B", 2},
		{"C", 4},
	}
	for _, tc := range cases {
		n, ok := nodes[tc.label]
		if !ok {
			t.Errorf("node %q not found", tc.label)
			continue
		}
		if n.In == nil {
			t.Errorf("node %q: In == nil", tc.label)
			continue
		}
		if n.In.Length != tc.want {
			t.Errorf("node %q: branch length = %g, want %g", tc.label, n.In.Length, tc.want)
		}
	}
	// Check the inner node branch length
	var innerNode *Node
	for _, n := range tree.Nodes {
		if n.IsInner() && n != tree.Root {
			innerNode = n
			break
		}
	}
	if innerNode == nil {
		t.Fatal("inner non-root node not found")
	}
	if innerNode.In.Length != 3 {
		t.Errorf("inner node branch length = %g, want 3", innerNode.In.Length)
	}
}

// TestReadNewick_MissingBranchLengths verifies that absent branch lengths are NaN.
func TestReadNewick_MissingBranchLengths(t *testing.T) {
	tree, _, err := readNewickString("(A,B);")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, n := range tree.Nodes {
		if n.In == nil {
			continue
		}
		if !math.IsNaN(n.In.Length) {
			t.Errorf("node %q: branch length = %g, want NaN", n.Label, n.In.Length)
		}
	}
}

// TestReadNewick_NodeCount verifies the total number of nodes in the tree.
func TestReadNewick_NodeCount(t *testing.T) {
	cases := []struct {
		input   string
		nbNodes int
	}{
		{"A;", 1},
		{"(A,B);", 3},
		{"((A,B),C);", 5},
		{"(A,B,C);", 4},
	}
	for _, tc := range cases {
		tree, _, err := readNewickString(tc.input)
		if err != nil {
			t.Errorf("%q: unexpected error: %v", tc.input, err)
			continue
		}
		if len(tree.Nodes) != tc.nbNodes {
			t.Errorf("%q: len(Nodes) = %d, want %d", tc.input, len(tree.Nodes), tc.nbNodes)
		}
	}
}

// TestReadNewick_RootBranchLength verifies that a branch length on the root is
// silently discarded (no error, root.In remains nil).
func TestReadNewick_RootBranchLength(t *testing.T) {
	tree, _, err := readNewickString("(A,B):0.5;")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tree.Root.In != nil {
		t.Errorf("root.In != nil after root branch length")
	}
}

// TestReadNewick_KnownTaxset verifies the TaxonSet.ReadNewick method with a
// pre-existing taxon set.
func TestReadNewick_KnownTaxset(t *testing.T) {
	taxset := makeTaxonSet(3) // A, B, C
	tree, err := readNewickStringWithTaxset("(A,B,C);", taxset, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodes := nodesByLabel(tree)
	for _, label := range []string{"A", "B", "C"} {
		n, ok := nodes[label]
		if !ok {
			t.Errorf("node %q not found", label)
			continue
		}
		wantId, _ := taxset.GetId(label)
		if n.TaxonId != wantId {
			t.Errorf("node %q: TaxonId = %d, want %d", label, n.TaxonId, wantId)
		}
	}
}

// TestReadNewick_AddNewFalse verifies that unknown taxa are not added to the
// taxon set when addNew=false, leaving TaxonId=-1.
func TestReadNewick_AddNewFalse(t *testing.T) {
	taxset := makeTaxonSet(2) // A, B only
	tree, err := readNewickStringWithTaxset("(A,B,C);", taxset, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if taxset.Len() != 2 {
		t.Errorf("taxset grew to %d, want 2", taxset.Len())
	}
	nodes := nodesByLabel(tree)
	c, ok := nodes["C"]
	if !ok {
		t.Fatal("node C not found")
	}
	if c.TaxonId != -1 {
		t.Errorf("C.TaxonId = %d, want -1", c.TaxonId)
	}
}

// TestReadNewick_Whitespace verifies that whitespace between tokens (after
// a symbol) is handled correctly.
func TestReadNewick_Whitespace(t *testing.T) {
	// Whitespace after a symbol (comma, parens) is consumed by the leading
	// whitespace loop in the tokenizer.
	cases := []string{
		"( A,B);",
		"(A, B);",
		"(A,\nB);",
		"( A , B ) ;",
	}
	for _, input := range cases {
		_, _, err := readNewickString(input)
		if err != nil {
			t.Errorf("%q: unexpected error: %v", input, err)
		}
	}
}

// TestReadNewick_Errors verifies that malformed Newick strings return errors.
func TestReadNewick_Errors(t *testing.T) {
	cases := []struct {
		input string
		desc  string
	}{
		{"((A,B);", "unmatched open paren"},
		{"(A,B)", "missing semicolon"},
		{"(A:xyz,B);", "invalid branch length"},
	}
	for _, tc := range cases {
		_, _, err := readNewickString(tc.input)
		if err == nil {
			t.Errorf("%s (%q): expected error, got nil", tc.desc, tc.input)
		}
	}
}