package trees

import (
	"fmt"
)

type Node struct {
	Id    int
	Label string
	In    *Branch
	Out   []*Branch
}

func (node *Node) String() string {
	if node.Label == "" {
		return fmt.Sprintf("Node #%d", node.Id)
	} else {
		return fmt.Sprintf("Node #%d - \"%s\"", node.Id, node.Label)
	}
}

type Branch struct {
	Id     int
	Parent *Node
	Child  *Node
	Length float64
}

func (branch *Branch) String() string {
	var parentSymbol, childSymbol string
	if branch.Parent == nil {
		parentSymbol = "◌"
	} else {
		parentSymbol = "●"
	}

	return fmt.Sprintf("Branch #%d "+parentSymbol+"───"+childSymbol+" (length: %f)", branch.Id, branch.Length)
}

type Tree struct {
	Root     *Node
	Nodes    []*Node
	Branches []*Branch
}

func (tree *Tree) String() string {
	return fmt.Sprintf("Tree (%d nodes)", tree.NbNodes())
}
