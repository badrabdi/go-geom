package geom

import "math"

type Geom0 struct {
	Lay     Layout
	Strd     int
	FlatCoord []float64
	Srid       int
}

type geom1 struct {
	Geom0
}

type geom2 struct {
	geom1
	ends []int
}

type geom3 struct {
	geom1
	endss [][]int
}

// Bounds returns the bounds of g.
func (g *Geom0) Bounds() *Bounds {
	return NewBounds(g.Lay).extendFlatCoords(g.FlatCoord, 0, len(g.FlatCoord), g.Strd)
}

// Coords returns all the coordinates in g, i.e. a single coordinate.
func (g *Geom0) Coords() Coord {
	return inflate0(g.FlatCoord, 0, len(g.FlatCoord), g.Strd)
}

// Empty returns true if g contains no coordinates.
func (g *Geom0) Empty() bool {
	return len(g.FlatCoord) == 0
}

// Ends returns the end indexes of sub-structures of g, i.e. an empty slice.
func (g *Geom0) Ends() []int {
	return nil
}

// Endss returns the end indexes of sub-sub-structures of g, i.e. an empty
// slice.
func (g *Geom0) Endss() [][]int {
	return nil
}

// FlatCoords returns the flat coordinates of g.
func (g *Geom0) FlatCoords() []float64 {
	return g.FlatCoord
}

// Layout returns g's Lay.
func (g *Geom0) Layout() Layout {
	return g.Lay
}

// NumCoords returns the number of coordinates in g, i.e. 1.
func (g *Geom0) NumCoords() int {
	return 1
}

// Reserve reserves space in g for n coordinates.
func (g *Geom0) Reserve(n int) {
	if cap(g.FlatCoord) < n*g.Strd {
		fcs := make([]float64, len(g.FlatCoord), n*g.Strd)
		copy(fcs, g.FlatCoord)
		g.FlatCoord = fcs
	}
}

// SRID returns g's SRID.
func (g *Geom0) SRID() int {
	return g.Srid
}

func (g *Geom0) setCoords(coords0 []float64) error {
	var err error
	g.FlatCoord, err = deflate0(nil, coords0, g.Strd)
	return err
}

// Stride returns g's Strd.
func (g *Geom0) Stride() int {
	return g.Strd
}

func (g *Geom0) verify() error {
	if g.Strd != g.Lay.Stride() {
		return errStrideLayoutMismatch
	}
	if g.Strd == 0 {
		if len(g.FlatCoord) != 0 {
			return errNonEmptyFlatCoords
		}
		return nil
	}
	if len(g.FlatCoord) != g.Strd {
		return errLengthStrideMismatch
	}
	return nil
}

// Coord returns the ith coord of g.
func (g *geom1) Coord(i int) Coord {
	return g.FlatCoord[i*g.Strd : (i+1)*g.Strd]
}

// Coords unpacks and returns all of g's coordinates.
func (g *geom1) Coords() []Coord {
	return inflate1(g.FlatCoord, 0, len(g.FlatCoord), g.Strd)
}

// NumCoords returns the number of coordinates in g.
func (g *geom1) NumCoords() int {
	return len(g.FlatCoord) / g.Strd
}

// Reverse reverses the order of g's coordinates.
func (g *geom1) Reverse() {
	reverse1(g.FlatCoord, 0, len(g.FlatCoord), g.Strd)
}

func (g *geom1) setCoords(coords1 []Coord) error {
	var err error
	g.FlatCoord, err = deflate1(nil, coords1, g.Strd)
	return err
}

func (g *geom1) verify() error {
	if g.Strd != g.Lay.Stride() {
		return errStrideLayoutMismatch
	}
	if g.Strd == 0 {
		if len(g.FlatCoord) != 0 {
			return errNonEmptyFlatCoords
		}
	} else {
		if len(g.FlatCoord)%g.Strd != 0 {
			return errLengthStrideMismatch
		}
	}
	return nil
}

// Coords returns all of g's coordinates.
func (g *geom2) Coords() [][]Coord {
	return inflate2(g.FlatCoord, 0, g.ends, g.Strd)
}

// Ends returns the end indexes of all sub-structures in g.
func (g *geom2) Ends() []int {
	return g.ends
}

// Reverse reverses the order of coordinates for each sub-structure in g.
func (g *geom2) Reverse() {
	reverse2(g.FlatCoord, 0, g.ends, g.Strd)
}

func (g *geom2) setCoords(coords2 [][]Coord) error {
	var err error
	g.FlatCoord, g.ends, err = deflate2(nil, nil, coords2, g.Strd)
	return err
}

func (g *geom2) verify() error {
	if g.Strd != g.Lay.Stride() {
		return errStrideLayoutMismatch
	}
	if g.Strd == 0 {
		if len(g.FlatCoord) != 0 {
			return errNonEmptyFlatCoords
		}
		if len(g.ends) != 0 {
			return errNonEmptyEnds
		}
		return nil
	}
	if len(g.FlatCoord)%g.Strd != 0 {
		return errLengthStrideMismatch
	}
	offset := 0
	for _, end := range g.ends {
		if end%g.Strd != 0 {
			return errMisalignedEnd
		}
		if end < offset {
			return errOutOfOrderEnd
		}
		offset = end
	}
	if offset != len(g.FlatCoord) {
		return errIncorrectEnd
	}
	return nil
}

// Coords returns all the coordinates in g.
func (g *geom3) Coords() [][][]Coord {
	return inflate3(g.FlatCoord, 0, g.endss, g.Strd)
}

// Endss returns a list of all the sub-sub-structures in g.
func (g *geom3) Endss() [][]int {
	return g.endss
}

// Reverse reverses the order of coordinates for each sub-sub-structure in g.
func (g *geom3) Reverse() {
	reverse3(g.FlatCoord, 0, g.endss, g.Strd)
}

func (g *geom3) setCoords(coords3 [][][]Coord) error {
	var err error
	g.FlatCoord, g.endss, err = deflate3(nil, nil, coords3, g.Strd)
	return err
}

func (g *geom3) verify() error {
	if g.Strd != g.Lay.Stride() {
		return errStrideLayoutMismatch
	}
	if g.Strd == 0 {
		if len(g.FlatCoord) != 0 {
			return errNonEmptyFlatCoords
		}
		if len(g.endss) != 0 {
			return errNonEmptyEndss
		}
		return nil
	}
	if len(g.FlatCoord)%g.Strd != 0 {
		return errLengthStrideMismatch
	}
	offset := 0
	for _, ends := range g.endss {
		for _, end := range ends {
			if end%g.Strd != 0 {
				return errMisalignedEnd
			}
			if end < offset {
				return errOutOfOrderEnd
			}
			offset = end
		}
	}
	if offset != len(g.FlatCoord) {
		return errIncorrectEnd
	}
	return nil
}

func doubleArea1(FlatCoord []float64, offset, end, Strd int) float64 {
	var doubleArea float64
	for i := offset + Strd; i < end; i += Strd {
		doubleArea += (FlatCoord[i+1] - FlatCoord[i+1-Strd]) * (FlatCoord[i] + FlatCoord[i-Strd])
	}
	return doubleArea
}

func doubleArea2(FlatCoord []float64, offset int, ends []int, Strd int) float64 {
	var doubleArea float64
	for i, end := range ends {
		da := doubleArea1(FlatCoord, offset, end, Strd)
		if i == 0 {
			doubleArea = da
		} else {
			doubleArea -= da
		}
		offset = end
	}
	return doubleArea
}

func doubleArea3(FlatCoord []float64, offset int, endss [][]int, Strd int) float64 {
	var doubleArea float64
	for _, ends := range endss {
		doubleArea += doubleArea2(FlatCoord, offset, ends, Strd)
		offset = ends[len(ends)-1]
	}
	return doubleArea
}

func deflate0(FlatCoord []float64, c Coord, Strd int) ([]float64, error) {
	if len(c) != Strd {
		return nil, ErrStrideMismatch{Got: len(c), Want: Strd}
	}
	FlatCoord = append(FlatCoord, c...)
	return FlatCoord, nil
}

func deflate1(FlatCoord []float64, coords1 []Coord, Strd int) ([]float64, error) {
	for _, c := range coords1 {
		var err error
		FlatCoord, err = deflate0(FlatCoord, c, Strd)
		if err != nil {
			return nil, err
		}
	}
	return FlatCoord, nil
}

func deflate2(
	FlatCoord []float64, ends []int, coords2 [][]Coord, Strd int,
) ([]float64, []int, error) {
	for _, coords1 := range coords2 {
		var err error
		FlatCoord, err = deflate1(FlatCoord, coords1, Strd)
		if err != nil {
			return nil, nil, err
		}
		ends = append(ends, len(FlatCoord))
	}
	return FlatCoord, ends, nil
}

func deflate3(
	FlatCoord []float64, endss [][]int, coords3 [][][]Coord, Strd int,
) ([]float64, [][]int, error) {
	for _, coords2 := range coords3 {
		var err error
		var ends []int
		FlatCoord, ends, err = deflate2(FlatCoord, ends, coords2, Strd)
		if err != nil {
			return nil, nil, err
		}
		endss = append(endss, ends)
	}
	return FlatCoord, endss, nil
}

func inflate0(FlatCoord []float64, offset, end, Strd int) Coord {
	if offset+Strd != end {
		panic("geom: Strd mismatch")
	}
	c := make([]float64, Strd)
	copy(c, FlatCoord[offset:end])
	return c
}

func inflate1(FlatCoord []float64, offset, end, Strd int) []Coord {
	coords1 := make([]Coord, (end-offset)/Strd)
	for i := range coords1 {
		coords1[i] = inflate0(FlatCoord, offset, offset+Strd, Strd)
		offset += Strd
	}
	return coords1
}

func inflate2(FlatCoord []float64, offset int, ends []int, Strd int) [][]Coord {
	coords2 := make([][]Coord, len(ends))
	for i := range coords2 {
		end := ends[i]
		coords2[i] = inflate1(FlatCoord, offset, end, Strd)
		offset = end
	}
	return coords2
}

func inflate3(FlatCoord []float64, offset int, endss [][]int, Strd int) [][][]Coord {
	coords3 := make([][][]Coord, len(endss))
	for i := range coords3 {
		ends := endss[i]
		coords3[i] = inflate2(FlatCoord, offset, ends, Strd)
		if len(ends) > 0 {
			offset = ends[len(ends)-1]
		}
	}
	return coords3
}

func length1(FlatCoord []float64, offset, end, Strd int) float64 {
	var length float64
	for i := offset + Strd; i < end; i += Strd {
		dx := FlatCoord[i] - FlatCoord[i-Strd]
		dy := FlatCoord[i+1] - FlatCoord[i+1-Strd]
		length += math.Sqrt(dx*dx + dy*dy)
	}
	return length
}

func length2(FlatCoord []float64, offset int, ends []int, Strd int) float64 {
	var length float64
	for _, end := range ends {
		length += length1(FlatCoord, offset, end, Strd)
		offset = end
	}
	return length
}

func length3(FlatCoord []float64, offset int, endss [][]int, Strd int) float64 {
	var length float64
	for _, ends := range endss {
		length += length2(FlatCoord, offset, ends, Strd)
		offset = ends[len(ends)-1]
	}
	return length
}

func reverse1(FlatCoord []float64, offset, end, Strd int) {
	for i, j := offset+Strd, end; i <= j; i, j = i+Strd, j-Strd {
		for k := 0; k < Strd; k++ {
			FlatCoord[i-Strd+k], FlatCoord[j-Strd+k] = FlatCoord[j-Strd+k], FlatCoord[i-Strd+k]
		}
	}
}

func reverse2(FlatCoord []float64, offset int, ends []int, Strd int) {
	for _, end := range ends {
		reverse1(FlatCoord, offset, end, Strd)
		offset = end
	}
}

func reverse3(FlatCoord []float64, offset int, endss [][]int, Strd int) {
	for _, ends := range endss {
		if len(ends) == 0 {
			continue
		}
		reverse2(FlatCoord, offset, ends, Strd)
		offset = ends[len(ends)-1]
	}
}
