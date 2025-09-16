package phylocore

import (
    "math"
    "testing"
)

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