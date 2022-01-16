package geom

// A GeometryCollection is a collection of arbitrary geometries with the same
// SRID.
type GeometryCollection struct {
	Lay Layout
	geoms  []T
	Srid   int
}

// NewGeometryCollection returns a new empty GeometryCollection.
func NewGeometryCollection() *GeometryCollection {
	return &GeometryCollection{}
}

// Geom returns the ith geometry in g.
func (g *GeometryCollection) Geom(i int) T {
	return g.geoms[i]
}

// Geoms returns the geometries in g.
func (g *GeometryCollection) Geoms() []T {
	return g.geoms
}

// Layout returns the smallest Lay that covers all of the layouts in g's
// geometries.
func (g *GeometryCollection) Layout() Layout {
	if g.Lay != NoLayout {
		return g.Lay
	}
	maxLayout := NoLayout
	for _, g := range g.geoms {
		switch l := g.Layout(); l {
		case XYZ:
			if maxLayout == XYM {
				maxLayout = XYZM
			} else if l > maxLayout {
				maxLayout = l
			}
		case XYM:
			if maxLayout == XYZ {
				maxLayout = XYZM
			} else if l > maxLayout {
				maxLayout = l
			}
		default:
			if l > maxLayout {
				maxLayout = l
			}
		}
	}
	return maxLayout
}

// NumGeoms returns the number of geometries in g.
func (g *GeometryCollection) NumGeoms() int {
	return len(g.geoms)
}

// Stride returns the Strd of g's Lay.
func (g *GeometryCollection) Stride() int {
	return g.Layout().Stride()
}

// Bounds returns the bounds of all the geometries in g.
func (g *GeometryCollection) Bounds() *Bounds {
	// FIXME this needs work for mixing layouts, e.g. XYZ and XYM
	b := NewBounds(g.Layout())
	for _, g := range g.geoms {
		b = b.Extend(g)
	}
	return b
}

// Empty returns true if the collection is empty.
// This can return true if the GeometryCollection contains multiple Geometry objects
// which are all empty.
func (g *GeometryCollection) Empty() bool {
	for _, g := range g.geoms {
		if !g.Empty() {
			return false
		}
	}
	return true
}

// FlatCoords panics.
func (g *GeometryCollection) FlatCoords() []float64 {
	panic("FlatCoords() called on a GeometryCollection")
}

// Ends panics.
func (g *GeometryCollection) Ends() []int {
	panic("Ends() called on a GeometryCollection")
}

// Endss panics.
func (g *GeometryCollection) Endss() [][]int {
	panic("Endss() called on a GeometryCollection")
}

// SRID returns g's SRID.
func (g *GeometryCollection) SRID() int {
	return g.Srid
}

// MustPush pushes gs to g. It panics on any error.
func (g *GeometryCollection) MustPush(gs ...T) *GeometryCollection {
	if err := g.Push(gs...); err != nil {
		panic(err)
	}
	return g
}

// CheckLayout checks all geometries in the collection match the given
// Lay.
func (g *GeometryCollection) CheckLayout(Lay Layout) error {
	if Lay != NoLayout {
		for _, geom := range g.geoms {
			if geomLayout := geom.Layout(); geomLayout != Lay {
				return ErrLayoutMismatch{
					Got:  Lay,
					Want: geomLayout,
				}
			}
		}
	}
	return nil
}

// MustSetLayout sets g's Lay. It panics on any error.
func (g *GeometryCollection) MustSetLayout(Lay Layout) *GeometryCollection {
	if err := g.SetLayout(Lay); err != nil {
		panic(err)
	}
	return g
}

// Push appends geometries.
func (g *GeometryCollection) Push(gs ...T) error {
	if g.Lay != NoLayout {
		for _, geom := range gs {
			if geomLayout := geom.Layout(); geomLayout != g.Lay {
				return ErrLayoutMismatch{
					Got:  geomLayout,
					Want: g.Lay,
				}
			}
		}
	}
	g.geoms = append(g.geoms, gs...)
	return nil
}

// SetLayout sets g's Lay.
func (g *GeometryCollection) SetLayout(Lay Layout) error {
	if err := g.CheckLayout(Lay); err != nil {
		return err
	}
	g.Lay = Lay
	return nil
}

// SetSRID sets g's SRID and the SRID of all its elements.
func (g *GeometryCollection) SetSRID(Srid int) *GeometryCollection {
	g.Srid = Srid
	return g
}
