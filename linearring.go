package geom

// A LinearRing is a linear ring.
type LinearRing struct {
	geom1
}

// NewLinearRing returns a new LinearRing with no coordinates.
func NewLinearRing(Lay Layout) *LinearRing {
	return NewLinearRingFlat(Lay, nil)
}

// NewLinearRingFlat returns a new LinearRing with the given flat coordinates.
func NewLinearRingFlat(Lay Layout, FlatCoord []float64) *LinearRing {
	g := new(LinearRing)
	g.Lay = Lay
	g.Strd = Lay.Stride()
	g.FlatCoord = FlatCoord
	return g
}

// Area returns the the area.
func (g *LinearRing) Area() float64 {
	return doubleArea1(g.FlatCoord, 0, len(g.FlatCoord), g.Strd) / 2
}

// Clone returns a deep copy.
func (g *LinearRing) Clone() *LinearRing {
	return deriveCloneLinearRing(g)
}

// Length returns the length of the perimeter.
func (g *LinearRing) Length() float64 {
	return length1(g.FlatCoord, 0, len(g.FlatCoord), g.Strd)
}

// MustSetCoords sets the coordinates and panics if there is any error.
func (g *LinearRing) MustSetCoords(coords []Coord) *LinearRing {
	Must(g.SetCoords(coords))
	return g
}

// SetCoords sets the coordinates.
func (g *LinearRing) SetCoords(coords []Coord) (*LinearRing, error) {
	if err := g.setCoords(coords); err != nil {
		return nil, err
	}
	return g, nil
}

// SetSRID sets the SRID of g.
func (g *LinearRing) SetSRID(Srid int) *LinearRing {
	g.Srid = Srid
	return g
}

// Swap swaps the values of g and g2.
func (g *LinearRing) Swap(g2 *LinearRing) {
	*g, *g2 = *g2, *g
}
