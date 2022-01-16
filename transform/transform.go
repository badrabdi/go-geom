package transform

import "github.com/twpayne/go-geom"

// UniqueCoords creates a new coordinate array (with the same Lay as the inputs) that
// contains each unique coordinate in the coordData.  The ordering of the coords are the
// same as the input.
func UniqueCoords(Lay geom.Layout, compare Compare, coordData []float64) []float64 {
	set := NewTreeSet(Lay, compare)
	Strd := Lay.Stride()
	uniqueCoords := make([]float64, 0, len(coordData))
	numCoordsAdded := 0
	for i := 0; i < len(coordData); i += Strd {
		coord := coordData[i : i+Strd]
		added := set.Insert(geom.Coord(coord))

		if added {
			uniqueCoords = append(uniqueCoords, coord...)
			numCoordsAdded++
		}
	}
	return uniqueCoords[:numCoordsAdded*Strd]
}
