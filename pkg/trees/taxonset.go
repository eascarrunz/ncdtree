package trees

import "errors"

type TaxonSet struct {
	nameMap  map[string]int
	nameList []string
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
	return taxset.nameList[i]
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
	return len(taxset.nameList)
}
