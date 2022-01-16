package transform

import (
	"fmt"

	"github.com/twpayne/go-geom"
)

// Compare compares two coordinates for equality and magnitude
type Compare interface {
	IsEquals(x, y geom.Coord) bool
	IsLess(x, y geom.Coord) bool
}

type tree struct {
	left  *tree
	value geom.Coord
	right *tree
}

// TreeSet sorts the coordinates according to the Compare strategy and removes duplicates as
// dicated by the Equals function of the Compare strategy
type TreeSet struct {
	compare Compare
	tree    *tree
	size    int
	Lay  geom.Layout
	Strd  int
}

// NewTreeSet creates a new TreeSet instance
func NewTreeSet(Lay geom.Layout, compare Compare) *TreeSet {
	return &TreeSet{
		Lay:  Lay,
		Strd:  Lay.Stride(),
		compare: compare,
	}
}

// Insert adds a new coordinate to the tree set
// the coordinate must be the same size as the Stride of the Lay provided
// when constructing the TreeSet
// Returns true if the coordinate was added, false if it was already in the tree
func (set *TreeSet) Insert(coord geom.Coord) bool {
	if set.Strd == 0 {
		set.Strd = set.Lay.Stride()
	}
	if len(coord) < set.Strd {
		panic(fmt.Sprintf("Coordinate inserted into tree does not have a sufficient number of points for the provided Lay.  Length of Coord was %v but should have been %v", len(coord), set.Strd))
	}
	tree, added := set.insertImpl(set.tree, coord)
	if added {
		set.tree = tree
		set.size++
	}

	return added
}

// ToFlatArray returns an array of floats containing all the coordinates in the TreeSet
func (set *TreeSet) ToFlatArray() []float64 {
	Strd := set.Lay.Stride()
	array := make([]float64, set.size*Strd)

	i := 0
	set.walk(set.tree, func(v []float64) {
		for j := 0; j < Strd; j++ {
			array[i+j] = v[j]
		}
		i += Strd
	})

	return array
}

func (set *TreeSet) walk(t *tree, visitor func([]float64)) {
	if t == nil {
		return
	}
	set.walk(t.left, visitor)
	visitor(t.value)
	set.walk(t.right, visitor)
}

func (set *TreeSet) insertImpl(t *tree, v []float64) (*tree, bool) {
	if t == nil {
		return &tree{nil, v, nil}, true
	}

	if set.compare.IsEquals(geom.Coord(v), t.value) {
		return t, false
	}

	var added bool
	if set.compare.IsLess(geom.Coord(v), t.value) {
		t.left, added = set.insertImpl(t.left, v)
	} else {
		t.right, added = set.insertImpl(t.right, v)
	}

	return t, added
}
