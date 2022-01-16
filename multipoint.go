package geom

// A MultiPoint is a collection of Points.
type MultiPoint struct {
	// To represent an MultiPoint that allows EMPTY elements, e.g.
	// MULTIPOINT ( EMPTY, POINT(1.0 1.0), EMPTY), we have to allow
	// record ends. If there is an empty point, ends[i] == ends[i-1].
	geom2
}

// NewMultiPoint returns a new, empty, MultiPoint.
func NewMultiPoint(Lay Layout) *MultiPoint {
	return NewMultiPointFlat(Lay, nil)
}

// NewMultiPointFlatOption represents an option that can be passed into
// NewMultiPointFlat.
type NewMultiPointFlatOption func(*MultiPoint)

// NewMultiPointFlatOptionWithEnds allows passing ends to NewMultiPointFlat,
// which allows the representation of empty points.
func NewMultiPointFlatOptionWithEnds(ends []int) NewMultiPointFlatOption {
	return func(mp *MultiPoint) {
		mp.ends = ends
	}
}

// NewMultiPointFlat returns a new MultiPoint with the given flat coordinates.
// Assumes no points are empty by default. Use `NewMultiPointFlatOptionWithEnds`
// to specify empty points.
func NewMultiPointFlat(
	Lay Layout, FlatCoord []float64, opts ...NewMultiPointFlatOption,
) *MultiPoint {
	g := new(MultiPoint)
	g.Lay = Lay
	g.Strd = Lay.Stride()
	g.FlatCoord = FlatCoord
	for _, opt := range opts {
		opt(g)
	}
	// If no ends are provided, assume all points are non empty.
	if g.ends == nil && len(g.FlatCoord) > 0 {
		numCoords := 0
		if g.Strd > 0 {
			numCoords = len(FlatCoord) / g.Strd
		}
		g.ends = make([]int, numCoords)
		for i := 0; i < numCoords; i++ {
			g.ends[i] = (i + 1) * g.Strd
		}
	}
	return g
}

// Area returns the area of g, i.e. zero.
func (g *MultiPoint) Area() float64 {
	return 0
}

// Clone returns a deep copy.
func (g *MultiPoint) Clone() *MultiPoint {
	return deriveCloneMultiPoint(g)
}

// Length returns zero.
func (g *MultiPoint) Length() float64 {
	return 0
}

// MustSetCoords sets the coordinates and panics on any error.
func (g *MultiPoint) MustSetCoords(coords []Coord) *MultiPoint {
	Must(g.SetCoords(coords))
	return g
}

// Coord returns the ith coord of g.
func (g *MultiPoint) Coord(i int) Coord {
	before := 0
	if i > 0 {
		before = g.ends[i-1]
	}
	if g.ends[i] == before {
		return nil
	}
	return g.FlatCoord[before:g.ends[i]]
}

// SetCoords sets the coordinates.
func (g *MultiPoint) SetCoords(coords []Coord) (*MultiPoint, error) {
	g.FlatCoord = nil
	g.ends = nil
	for _, c := range coords {
		if c != nil {
			var err error
			g.FlatCoord, err = deflate0(g.FlatCoord, c, g.Strd)
			if err != nil {
				return nil, err
			}
		}
		g.ends = append(g.ends, len(g.FlatCoord))
	}
	return g, nil
}

// Coords unpacks and returns all of g's coordinates.
func (g *MultiPoint) Coords() []Coord {
	coords1 := make([]Coord, len(g.ends))
	offset := 0
	prevEnd := 0
	for i, end := range g.ends {
		if end != prevEnd {
			coords1[i] = inflate0(g.FlatCoord, offset, offset+g.Strd, g.Strd)
			offset += g.Strd
		}
		prevEnd = end
	}
	return coords1
}

// NumCoords returns the number of coordinates in g.
func (g *MultiPoint) NumCoords() int {
	return len(g.ends)
}

// SetSRID sets the SRID of g.
func (g *MultiPoint) SetSRID(Srid int) *MultiPoint {
	g.Srid = Srid
	return g
}

// NumPoints returns the number of Points.
func (g *MultiPoint) NumPoints() int {
	return len(g.ends)
}

// Point returns the ith Point.
func (g *MultiPoint) Point(i int) *Point {
	coord := g.Coord(i)
	if coord == nil {
		return NewPointEmpty(g.Lay)
	}
	return NewPointFlat(g.Lay, coord)
}

// Push appends a point.
func (g *MultiPoint) Push(p *Point) error {
	if p.Lay != g.Lay {
		return ErrLayoutMismatch{Got: p.Lay, Want: g.Lay}
	}
	if !p.Empty() {
		g.FlatCoord = append(g.FlatCoord, p.FlatCoord...)
	}
	g.ends = append(g.ends, len(g.FlatCoord))
	return nil
}

// Swap swaps the values of g and g2.
func (g *MultiPoint) Swap(g2 *MultiPoint) {
	*g, *g2 = *g2, *g
}
