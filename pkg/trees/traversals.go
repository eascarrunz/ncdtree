package trees

type Traversal int

const (
	// Type of traversal where each node is visited before its children.
	PreOrder = iota
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

func (branch *Branch) Traverse(f func(*Branch), traversal Traversal) {
	switch traversal {
	case PreOrder:
		branch.preTraverse(f)
	case PostOrder:
		branch.postTraverse(f)
	}
}

func (tree *Tree) TraverseNodes(f func(*Node), traversal Traversal) {
	tree.Root.Traverse(f, traversal)
}

func (tree *Tree) TraverseBranches(f func(*Branch), traversal Traversal) {
	for _, branch := range tree.Root.Out {
		branch.Traverse(f, traversal)
	}
}

func (node *Node) NbDescendants() int {
	nb := 0
	count := func(*Node) {
		nb += 1
	}

	node.Traverse(count, PreOrder)

	return nb
}

func (tree *Tree) NbNodes() int {
	return tree.Root.NbDescendants()
}

func (tree *Tree) NbBranches() int {
	return tree.NbNodes() - 1
}
