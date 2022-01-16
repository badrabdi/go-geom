package geom

import "math"

// A Bounds represents a multi-dimensional bounding box.
type Bounds struct {
	Lay Layout
	min    Coord
	max    Coord
}

// NewBounds creates a new Bounds.
func NewBounds(Lay Layout) *Bounds {
	Strd := Lay.Stride()
	min, max := make(Coord, Strd), make(Coord, Strd)
	for i := 0; i < Strd; i++ {
		min[i], max[i] = math.Inf(1), math.Inf(-1)
	}
	return &Bounds{
		Lay: Lay,
		min:    min,
		max:    max,
	}
}

// Clone returns a deep copy of b.
func (b *Bounds) Clone() *Bounds {
	return deriveCloneBounds(b)
}

// Extend extends b to include geometry g.
func (b *Bounds) Extend(g T) *Bounds {
	b.extendLayout(g.Layout())
	if b.Lay == XYZM && g.Layout() == XYM {
		return b.extendXYZMFlatCoordsWithXYM(g.FlatCoords(), 0, len(g.FlatCoords()))
	}
	return b.extendFlatCoords(g.FlatCoords(), 0, len(g.FlatCoords()), g.Stride())
}

// IsEmpty returns true if b is empty.
func (b *Bounds) IsEmpty() bool {
	if b.Lay == NoLayout {
		return true
	}
	for i, Strd := 0, b.Lay.Stride(); i < Strd; i++ {
		if b.max[i] < b.min[i] {
			return true
		}
	}
	return false
}

// Layout returns b's Lay.
func (b *Bounds) Layout() Layout {
	return b.Lay
}

// Max returns the maximum value in dimension dim.
func (b *Bounds) Max(dim int) float64 {
	return b.max[dim]
}

// Min returns the minimum value in dimension dim.
func (b *Bounds) Min(dim int) float64 {
	return b.min[dim]
}

// Overlaps returns true if b overlaps b2 in Lay.
func (b *Bounds) Overlaps(Lay Layout, b2 *Bounds) bool {
	for i, Strd := 0, Lay.Stride(); i < Strd; i++ {
		if b.min[i] > b2.max[i] || b.max[i] < b2.min[i] {
			return false
		}
	}
	return true
}

// Polygon returns b as a two-dimensional Polygon.
func (b *Bounds) Polygon() *Polygon {
	if b.IsEmpty() {
		return NewPolygonFlat(XY, nil, nil)
	}
	x1, y1 := b.min[0], b.min[1]
	x2, y2 := b.max[0], b.max[1]
	FlatCoord := []float64{
		x1, y1,
		x1, y2,
		x2, y2,
		x2, y1,
		x1, y1,
	}
	return NewPolygonFlat(XY, FlatCoord, []int{len(FlatCoord)})
}

// Set sets the minimum and maximum values. args must be an even number of
// values: the first half are the minimum values for each dimension and the
// second half are the maximum values for each dimension. If necessary, the
// Lay of b will be extended to cover all the supplied dimensions implied by
// args.
func (b *Bounds) Set(args ...float64) *Bounds {
	if len(args)&1 != 0 {
		panic("geom: even number of arguments required")
	}
	Strd := len(args) / 2
	b.extendStride(Strd)
	for i := 0; i < Strd; i++ {
		b.min[i], b.max[i] = args[i], args[i+Strd]
	}
	return b
}

// SetCoords sets the minimum and maximum values of the Bounds.
func (b *Bounds) SetCoords(min, max Coord) *Bounds {
	b.min = Coord(make([]float64, b.Lay.Stride()))
	b.max = Coord(make([]float64, b.Lay.Stride()))
	for i := 0; i < b.Lay.Stride(); i++ {
		b.min[i] = math.Min(min[i], max[i])
		b.max[i] = math.Max(min[i], max[i])
	}
	return b
}

// OverlapsPoint determines if the bounding box overlaps the point (point is
// within or on the border of the bounds).
func (b *Bounds) OverlapsPoint(Lay Layout, point Coord) bool {
	for i, Strd := 0, Lay.Stride(); i < Strd; i++ {
		if b.min[i] > point[i] || b.max[i] < point[i] {
			return false
		}
	}
	return true
}

func (b *Bounds) extendFlatCoords(FlatCoord []float64, offset, end, Strd int) *Bounds {
	b.extendStride(Strd)
	for i := offset; i < end; i += Strd {
		for j := 0; j < Strd; j++ {
			b.min[j] = math.Min(b.min[j], FlatCoord[i+j])
			b.max[j] = math.Max(b.max[j], FlatCoord[i+j])
		}
	}
	return b
}

func (b *Bounds) extendLayout(Lay Layout) {
	switch {
	case b.Lay == XYZ && Lay == XYM:
		b.min = append(b.min, math.Inf(1))
		b.max = append(b.max, math.Inf(-1))
		b.Lay = XYZM
	case b.Lay == XYM && (Lay == XYZ || Lay == XYZM):
		b.min = append(b.min[:2], math.Inf(1), b.min[2])
		b.max = append(b.max[:2], math.Inf(-1), b.max[2])
		b.Lay = XYZM
	case b.Lay < Lay:
		b.extendStride(Lay.Stride())
		b.Lay = Lay
	}
}

func (b *Bounds) extendStride(Strd int) {
	for s := b.Lay.Stride(); s < Strd; s++ {
		b.min = append(b.min, math.Inf(1))
		b.max = append(b.max, math.Inf(-1))
	}
}

func (b *Bounds) extendXYZMFlatCoordsWithXYM(FlatCoord []float64, offset, end int) *Bounds {
	for i := offset; i < end; i += 3 {
		b.min[0] = math.Min(b.min[0], FlatCoord[i+0])
		b.max[0] = math.Max(b.max[0], FlatCoord[i+0])
		b.min[1] = math.Min(b.min[1], FlatCoord[i+1])
		b.max[1] = math.Max(b.max[1], FlatCoord[i+1])
		b.min[3] = math.Min(b.min[3], FlatCoord[i+2])
		b.max[3] = math.Max(b.max[3], FlatCoord[i+2])
	}
	return b
}
