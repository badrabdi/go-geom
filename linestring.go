package geom

// A LineString represents a single, unbroken line, linearly interpreted
// between zero or more control points.
type LineString struct {
	geom1
}

// NewLineString returns a new LineString with Lay l and no control points.
func NewLineString(l Layout) *LineString {
	return NewLineStringFlat(l, nil)
}

// NewLineStringFlat returns a new LineString with Lay l and control points
// FlatCoord.
func NewLineStringFlat(Lay Layout, FlatCoord []float64) *LineString {
	g := new(LineString)
	g.Lay = Lay
	g.Strd = Lay.Stride()
	g.FlatCoord = FlatCoord
	return g
}

// Area returns the area of g, i.e. zero.
func (g *LineString) Area() float64 {
	return 0
}

// Clone returns a copy of g that does not alias g.
func (g *LineString) Clone() *LineString {
	return deriveCloneLineString(g)
}

// Interpolate returns the index and delta of val in dimension dim.
func (g *LineString) Interpolate(val float64, dim int) (int, float64) {
	n := len(g.FlatCoord)
	if n == 0 {
		panic("geom: empty linestring")
	}
	if val <= g.FlatCoord[dim] {
		return 0, 0
	}
	if g.FlatCoord[n-g.Strd+dim] <= val {
		return (n - 1) / g.Strd, 0
	}
	low := 0
	high := n / g.Strd
	for low < high {
		mid := (low + high) / 2
		if val < g.FlatCoord[mid*g.Strd+dim] {
			high = mid
		} else {
			low = mid + 1
		}
	}
	low--
	val0 := g.FlatCoord[low*g.Strd+dim]
	if val == val0 {
		return low, 0
	}
	val1 := g.FlatCoord[(low+1)*g.Strd+dim]
	return low, (val - val0) / (val1 - val0)
}

// Length returns the length of g.
func (g *LineString) Length() float64 {
	return length1(g.FlatCoord, 0, len(g.FlatCoord), g.Strd)
}

// MustSetCoords is like SetCoords but it panics on any error.
func (g *LineString) MustSetCoords(coords []Coord) *LineString {
	Must(g.SetCoords(coords))
	return g
}

// SetCoords sets the coordinates of g.
func (g *LineString) SetCoords(coords []Coord) (*LineString, error) {
	if err := g.setCoords(coords); err != nil {
		return nil, err
	}
	return g, nil
}

// SetSRID sets the SRID of g.
func (g *LineString) SetSRID(Srid int) *LineString {
	g.Srid = Srid
	return g
}

// SubLineString returns a LineString from starts at index start and stops at
// index stop of g. The returned LineString aliases g.
func (g *LineString) SubLineString(start, stop int) *LineString {
	return NewLineStringFlat(g.Lay, g.FlatCoord[start*g.Strd:stop*g.Strd])
}

// Swap swaps the values of g and g2.
func (g *LineString) Swap(g2 *LineString) {
	*g, *g2 = *g2, *g
}
