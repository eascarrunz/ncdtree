package phylocore

type Traversal int

const (
	// Type of traversal where each node is visited before its children.
	PreOrder Traversal = iota
	// Type of traversal where each node is visited after its children.
	PostOrder
)

func (node *Node) preTraverse(f func(*Node)) {
	f(node)

	for _, branch := range node.Out {
		branch.Child.preTraverse(f)
	}
}

func (node *Node) postTraverse(f func(*Node)) {
	for _, branch := range node.Out {
		branch.Child.postTraverse(f)
	}

	f(node)
}

/*
Apply the provided function f to the nodes of the sub-tree descendant from the current node, in the specified traversal order.

# Parameters
  - f: A function of signature func(*Node)
  - traversal: Type of traversal order (PreOrder or PostOrder)
*/
func (node *Node) Traverse(f func(*Node), traversal Traversal) {
	switch traversal {
	case PreOrder:
		node.preTraverse(f)
	case PostOrder:
		node.postTraverse(f)
	}
}

func (branch *Branch) preTraverse(f func(*Branch)) {
	f(branch)

	for _, childBranch := range branch.Child.Out {
		childBranch.preTraverse(f)
	}
}

func (branch *Branch) postTraverse(f func(*Branch)) {
	for _, childBranch := range branch.Child.Out {
		childBranch.preTraverse(f)
	}

	f(branch)
}

/*
Apply the provided function f to the branches of the sub-tree descendant from the current branch, in the specified traversal order

# Parameters
  - f: A function of signature func(*Branch)
  - traversal: Type of traversal order (PreOrder or PostOrder)
*/
func (branch *Branch) Traverse(f func(*Branch), traversal Traversal) {
	switch traversal {
	case PreOrder:
		branch.preTraverse(f)
	case PostOrder:
		branch.postTraverse(f)
	}
}

/*
Apply the provided function f to the nodes of the tree in the specified traversal order

# Parameters
  - f: A function of signature func(*Node)
  - traversal: Type of traversal order (PreOrder or PostOrder)
*/
func (tree *Tree) TraverseNodes(f func(*Node), traversal Traversal) {
	tree.Root.Traverse(f, traversal)
}

/*
Apply the provided function f to the branches of the tree in the specified traversal order

# Parameters
  - f: A function of signature func(*Branch)
  - traversal: Type of traversal order (PreOrder or PostOrder)
*/
func (tree *Tree) TraverseBranches(f func(*Branch), traversal Traversal) {
	for _, branch := range tree.Root.Out {
		branch.Traverse(f, traversal)
	}
}

/*
Return the number of nodes that descend from the current node
*/
func (node *Node) NbDescendants() int {
	nb := 0
	count := func(*Node) {
		nb += 1
	}

	node.Traverse(count, PreOrder)

	return nb
}

/*
Return the number of nodes in the tree
*/
func (tree *Tree) NbNodes() int {
	return tree.Root.NbDescendants()
}

/*
Return the number of branches in the tree
*/
func (tree *Tree) NbBranches() int {
	return tree.NbNodes() - 1
}
