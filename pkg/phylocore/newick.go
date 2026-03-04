package phylocore

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

/*
Returns the Newick representation of the subtree rooted in the node, omitting the final semicolon.
*/
func (node *Node) makeNewick(b *strings.Builder) {
	if len(node.Out) > 0 {
		b.WriteString("(")
		isFirst := true
		for _, branch := range node.Out {
			if !isFirst {
				b.WriteString(",")
			}

			branch.Child.makeNewick(b)

			if !math.IsNaN(branch.Length) {
				b.WriteString(":")
				b.WriteString(strconv.FormatFloat(branch.Length, 'g', 6, 64))
			}
			isFirst = false
		}
		b.WriteString(")")
	}
	b.WriteString(node.Label)
}

/*
Returns the Newick representation of the subtree rooted in the node.

Reference for the Newick format: https://phylipweb.github.io/phylip/newicktree.html
*/
func (node *Node) NewickString() string {
	var b strings.Builder
	node.makeNewick(&b)
	b.WriteString(";")

	return b.String()
}

/*
Returns the Newick representation of the tree.

References for the Newick format:
  - https://phylipweb.github.io/phylip/newicktree.html
  - https://en.wikipedia.org/wiki/Newick_format
*/
func (tree *Tree) NewickString() string {
	return tree.Root.NewickString()
}

type newickToken int

const (
	tknNil newickToken = iota
	tknTerminate
	tknOpenParens
	tknCloseParens
	tknComma
	tknColon
	tknValue // For raw strings, both labels and branch lengths
)

type newickTokenizer struct {
	stream  *bufio.Reader
	builder *strings.Builder
	token   newickToken
	value   string
}

func recognizeNewickToken(c rune) newickToken {
	switch c {
	case ';':
		return tknTerminate
	case '(':
		return tknOpenParens
	case ')':
		return tknCloseParens
	case ',':
		return tknComma
	case ':':
		return tknColon
	default:
		return tknValue
	}
}

func (tokenizer *newickTokenizer) Read() {
	c, _, err := tokenizer.stream.ReadRune()
	if err != nil {
		panic(fmt.Sprintf("error reading Newick string: %v", err))
	}

	// Consume whitespace
	for unicode.IsSpace(c) {
		c, _, err = tokenizer.stream.ReadRune()
		if err != nil {
			panic(fmt.Sprintf("error reading Newick string: %v", err))
		}
	}

	// Try to read Newick symbol
	tkn := recognizeNewickToken(c)
	if tkn != tknValue {
		tokenizer.token = tkn
		tokenizer.value = ""

		return
	}

	// Read value
	tokenizer.builder.Reset()

	for tkn == tknValue {
		if unicode.IsSpace(c) {
			break
		}
		tokenizer.builder.WriteRune(c)

		c, _, err = tokenizer.stream.ReadRune()
		if err != nil {
			panic(fmt.Sprintf("error reading Newick string: %v", err))
		}

		tkn = recognizeNewickToken(c)
	}
	tokenizer.token = tknValue
	tokenizer.value = tokenizer.builder.String()

	// Next token is a Newick symbol, unread the rune for the next tokenizer call
	tokenizer.stream.UnreadRune()
}

type newickParseContext struct {
	taxset        *TaxonSet
	tree          *Tree
	taxonTargets  NodeGroup
	acceptNewTaxa bool
}

func (ctx *newickParseContext) setTaxon(node *Node) {
	node.TaxonId = -1

	switch ctx.taxonTargets {
	case OuterNodes:
		if node.IsInner() {
			return
		}
	case InnerNodes:
		if node.IsOuter() {
			return
		}
	}

	taxonId, ok := ctx.taxset.GetId(node.Label)
	if !ok {
		if ctx.acceptNewTaxa {
			taxonId = ctx.taxset.NewTaxon(node.Label)
			node.TaxonId = taxonId
		}
	} else {
		node.TaxonId = taxonId
	}
}

func (taxset *TaxonSet) parseNewick(reader *bufio.Reader, addNew bool) (tree *Tree, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("newick: %v", r)
		}
	}()

	if taxset.Len() > 0 {
		tree = NewEmptyTree(2*taxset.Len() - 1)
	} else {
		tree = NewEmptyTree(0)
	}

	var builder strings.Builder
	ctx := newickParseContext{
		taxset, tree, OuterNodes, addNew,
	}
	tokenizer := newickTokenizer{
		reader, &builder, tknNil, "",
	}

	tokenizer.Read()

	ctx.parseRoot(&tokenizer)

	if tokenizer.token != tknTerminate {
		panic("Newick not terminated with ';'")
	}

	return tree, nil
}

// Read a Newick string and return a tree with a matching taxon set
func ReadNewick(reader *bufio.Reader) (*Tree, *TaxonSet, error) {
	taxset, _ := NewTaxonSet(make([]string, 0))
	tree, err := taxset.parseNewick(reader, true)

	return tree, taxset, err
}

// Read a Newick string and return a tree with taxa matching a given taxon set, optionally adding new taxa
func (taxset *TaxonSet) ReadNewick(reader *bufio.Reader, addNew bool) (*Tree, error) {
	return taxset.parseNewick(reader, addNew)
}

func parseBranchLength(node *Node, tokenizer *newickTokenizer) {
	tokenizer.Read()

	if tokenizer.token == tknValue {
		brlength, err := strconv.ParseFloat(tokenizer.value, 64)
		if err != nil {
			panic(fmt.Sprintf("Could not parse branch length: \"%s\"", tokenizer.value))
		}

		// Edge case: Root with branch length
		if node.In == nil {
			fmt.Fprint(os.Stderr, "branch length in root discarded\n")
		} else {
			node.In.Length = brlength
		}

		// Get next token
		tokenizer.Read()
	}
	// If no branch length value is given, keep length as NaN
}

/* Handles edge case of the root being a node without children. */
func (ctx *newickParseContext) parseRoot(tokenizer *newickTokenizer) {
	root := ctx.tree.NewNode()
	ctx.tree.Root = root
	if tokenizer.token == tknOpenParens {
		ctx.parseInnerNode(root, tokenizer)
	} else {
		ctx.parseOuterNode(root, tokenizer)
	}
}

func (ctx *newickParseContext) parseOuterNode(node *Node, tokenizer *newickTokenizer) {
	if tokenizer.token == tknValue {
		node.Label = tokenizer.value
		ctx.setTaxon(node)

		tokenizer.Read()
	}

	// Branch length next?
	if tokenizer.token == tknColon {
		parseBranchLength(node, tokenizer)
	}
}

func (ctx *newickParseContext) parseInnerNode(node *Node, tokenizer *newickTokenizer) {
	// Consume the open parens
	if tokenizer.token != tknOpenParens {
		panic("Expected '(' in Newick string")
	}
	tokenizer.Read()

	// Read the list of children
	for {
		if tokenizer.token == tknOpenParens {
			// Is an inner node
			child := ctx.tree.NewNode()
			node.AddChild(child, ctx.tree.NewBranch())
			ctx.parseInnerNode(child, tokenizer)
		} else {
			// Is an outer node
			child := ctx.tree.NewNode()
			node.AddChild(child, ctx.tree.NewBranch())
			ctx.parseOuterNode(child, tokenizer)
		}

		if tokenizer.token != tknComma {
			break
		}
		tokenizer.Read()
	}

	// Now the current token must be a closing parens
	if tokenizer.token != tknCloseParens {
		panic("Unmatched '(' in Newick string")
	}

	tokenizer.Read()

	if tokenizer.token == tknValue {
		node.Label = tokenizer.value
		ctx.setTaxon(node)

		tokenizer.Read()
	}

	// Branch length next?
	if tokenizer.token == tknColon {
		parseBranchLength(node, tokenizer)
	}
}
