package phylocore

import (
	"errors"
	"fmt"
)

/*
A structure for handling a list of taxa represented by unique names associated to unique numeric IDs.
*/
type TaxonSet struct {
	nameMap map[string]int
	Names   []string
}

func (taxset TaxonSet) String() string {
	return fmt.Sprintf("TaxonSet (%d taxa)", taxset.Len())
}

// Create a new taxon set based on a list of names
func NewTaxonSet(nameList []string) (*TaxonSet, error) {
	nameMap := make(map[string]int)

	for i, name := range nameList {
		_, err := nameMap[name]
		if !err {
			nameMap[name] = i
		} else {
			return nil, errors.New("duplicate name \"" + name + "\"")
		}
	}

	taxset := TaxonSet{nameMap, nameList}

	return &taxset, nil
}

// Get the name of a taxon by its numeric ID
func (taxset *TaxonSet) GetName(i int) string {
	return taxset.Names[i]
}

// Get the numeric ID of a taxon by its name
func (taxset *TaxonSet) GetId(s string) int {
	id, ok := taxset.nameMap[s]
	if !ok {
		panic("name \"" + s + "\" does not exist in taxon set")
	}

	return id
}

// Return the length of a taxon set
func (taxset *TaxonSet) Len() int {
	return len(taxset.Names)
}

func (taxset *TaxonSet) MakeUnassembledTree() *Tree {
	nbOuter := taxset.Len()
	nbNode := 2*nbOuter - 1
	tree := MakeUnassembledTree(nbNode)

	for i := range nbOuter {
		tree.Nodes[i].TaxonId = i
		tree.Nodes[i].Label = taxset.GetName(i)
	}

	tree.Root = tree.Nodes[nbOuter]

	return tree
}
