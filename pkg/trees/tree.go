package trees

import (
	"fmt"
	"math"
)

type Node struct {
	Id      int
	TaxonId int
	Label   string
	In      *Branch
	Out     []*Branch
}

func (node *Node) String() string {
	if node.Label == "" {
		return fmt.Sprintf("Node #%d", node.Id)
	} else {
		return fmt.Sprintf("Node #%d - \"%s\"", node.Id, node.Label)
	}
}

/*
Return the number of parents of the node (0 for the root, 1 for all others)
*/
func (node *Node) InDegree() int {
	if node.In == nil {
		return 0
	} else {
		return 1
	}
}

/*
Return the number of children of the node
*/
func (node *Node) OutDegree() int {
	return len(node.Out)
}

/*
Return the number of neighbours of the node (parent + children)
*/
func (node *Node) Degree() int {
	return node.InDegree() + node.OutDegree()
}

/*
Check whether the node is "inner" (also called "internal"), i.e. it has children
*/
func (node *Node) IsInner() bool {
	return node.OutDegree() > 0
}

/*
Check whether the node is "outer" (also "external", "tip" or "leaf"), i.e. it has no children
*/
func (node *Node) IsOuter() bool {
	return !node.IsInner()
}

type Branch struct {
	Id     int
	Parent *Node
	Child  *Node
	Length float64
}

func (branch *Branch) String() string {
	var parentString, childString string

	if branch.Parent == nil {
		parentString = "◌"
	} else {
		parentString = fmt.Sprintf("%d●", branch.Parent.Id)
	}

	if branch.Child == nil {
		childString = "◌"
	} else {
		childString = fmt.Sprintf("●%d", branch.Child.Id)
	}

	return fmt.Sprintf("Branch #%d (length %f): "+parentString+"——⟶"+childString, branch.Id, branch.Length)
}

type Tree struct {
	Root     *Node
	Nodes    []*Node
	Branches []*Branch
}

func (tree *Tree) String() string {
	return fmt.Sprintf("Tree (%d nodes)", tree.NbNodes())
}

/*
Create a tree object with initialized nodes.

The nodes are initialized with individual node IDs and -1 as the taxon ID (meaning "no taxon").

# Arguments
  - nbNode: Number of nodes to initialize
*/
func MakeUnassembledTree(nbNode int) *Tree {
	nbBranch := nbNode - 1
	nodes := make([]*Node, nbNode)
	branches := make([]*Branch, 0, nbBranch)

	for i := range nbNode {
		nodes[i] = &Node{i, -1, "", nil, make([]*Branch, 0, 3)}
	}

	return &Tree{
		Root:     nil,
		Nodes:    nodes,
		Branches: branches,
	}
}

/*
Create a new branch in the tree. Use this function to ensure a valid ID is assigned to the branch.
*/
func (tree *Tree) NewBranch() *Branch {
	newBranch := &Branch{len(tree.Branches), nil, nil, math.NaN()}
	tree.Branches = append(tree.Branches, newBranch)

	return newBranch
}

/*
Create a new node in the tree. Use this function to ensure a valid ID is assigned to the node.
*/
func (tree *Tree) NewNode() *Node {
	newNode := &Node{len(tree.Nodes), -1, "", nil, make([]*Branch, 0, 3)}
	tree.Nodes = append(tree.Nodes, newNode)

	return newNode
}

/*
Link up a parent node to a child node via a given branch.
*/
func (parent *Node) AddChild(child *Node, branch *Branch) {
	branch.Parent = parent
	branch.Child = child
	parent.Out = append(parent.Out, branch)
	child.In = branch
}

/*
Sever the connections between a branch and its child node. The parent node remains attached.
*/
func (branch *Branch) SeparateChild() {
	branch.Child.In = nil
	branch.Child = nil
}

/*
Sever the connections between a branch and its parent node. The child node remains attached.
*/
func (branch *Branch) SeparateParent() {
	var childIndex int
	for i, childBranch := range branch.Parent.Out {
		if childBranch == branch {
			childIndex = i
			break
		}
	}

	copy(branch.Parent.Out[childIndex:], branch.Parent.Out[childIndex+1:])
	branch.Parent.Out[len(branch.Parent.Out)-1] = nil
	branch.Parent.Out = branch.Parent.Out[:len(branch.Parent.Out)-1]

	branch.Parent = nil
}

/*
Sever the connections between a branch and its parent and child nodes.
*/
func (branch *Branch) Separate() {
	branch.SeparateChild()
	branch.SeparateParent()
}
