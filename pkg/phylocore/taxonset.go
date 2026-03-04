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

// Add a new taxon to the taxon set and return its ID.
func (taxset *TaxonSet) NewTaxon(name string) int {
	i := len(taxset.Names)
	_, err := taxset.nameMap[name]
	if !err {
		taxset.nameMap[name] = i
		taxset.Names = append(taxset.Names, name)
	} else {
		panic(fmt.Sprintf("duplicate name \"%s\"", name))
	}

	return i
}

// Get the name of a taxon by its numeric ID
func (taxset *TaxonSet) GetName(i int) (string, bool) {
	if i > taxset.Len() {
		return "", false
	}
	return taxset.Names[i], true
}

// Get the numeric ID of a taxon by its name
func (taxset *TaxonSet) GetId(s string) (int, bool) {
	id, ok := taxset.nameMap[s]

	return id, ok
}

// Return the length of a taxon set
func (taxset *TaxonSet) Len() int {
	return len(taxset.Names)
}

func (taxset *TaxonSet) MakeUnassembledTree() *Tree {
	nbOuter := taxset.Len()
	nbNode := 2*nbOuter - 1
	tree := MakeUnassembledTreePhylip(nbNode)

	for i := range nbOuter {
		tree.Nodes[i].TaxonId = i
		label, _ := taxset.GetName(i)
		tree.Nodes[i].Label = label
	}

	tree.Root = tree.Nodes[nbOuter]

	return tree
}
