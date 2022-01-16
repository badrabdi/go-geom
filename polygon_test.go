package geom

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Polygon implements interface T.
var _ T = &Polygon{}

func ExampleNewPolygon() {
	unitSquare := NewPolygon(XY).MustSetCoords([][]Coord{
		{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}},
	})
	fmt.Printf("unitSquare.Area() == %f", unitSquare.Area())
	// Output: unitSquare.Area() == 1.000000
}

type expectedPolygon struct {
	Lay     Layout
	Strd     int
	FlatCoord []float64
	ends       []int
	coords     [][]Coord
	bounds     *Bounds
}

func (g *Polygon) assertEquals(t *testing.T, e *expectedPolygon) {
	t.Helper()
	assert.NoError(t, g.verify())
	assert.Equal(t, e.Lay, g.Layout())
	assert.Equal(t, e.Strd, g.Stride())
	assert.Equal(t, e.FlatCoord, g.FlatCoords())
	assert.Equal(t, e.ends, g.Ends())
	assert.Nil(t, g.Endss())
	assert.Equal(t, e.coords, g.Coords())
	assert.Equal(t, e.bounds, g.Bounds())
	assert.Equal(t, len(e.coords), g.NumLinearRings())
	for i, c := range e.coords {
		assert.Equal(t, NewLinearRing(g.Layout()).MustSetCoords(c), g.LinearRing(i))
	}
}

func TestPolygon(t *testing.T) {
	for i, tc := range []struct {
		p        *Polygon
		expected *expectedPolygon
	}{
		{
			p: NewPolygon(XY).MustSetCoords([][]Coord{{{1, 2}, {3, 4}, {5, 6}}, {{7, 8}, {9, 10}, {11, 12}}}),
			expected: &expectedPolygon{
				Lay:     XY,
				Strd:     2,
				coords:     [][]Coord{{{1, 2}, {3, 4}, {5, 6}}, {{7, 8}, {9, 10}, {11, 12}}},
				FlatCoord: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				ends:       []int{6, 12},
				bounds:     NewBounds(XY).Set(1, 2, 11, 12),
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tc.p.assertEquals(t, tc.expected)
			assert.False(t, aliases(tc.p.FlatCoords(), tc.p.Clone().FlatCoords()))
		})
	}
}

func TestPolygonStrideMismatch(t *testing.T) {
	for i, tc := range []struct {
		l        Layout
		cs       [][]Coord
		expected error
	}{
		{
			l:        XY,
			cs:       nil,
			expected: nil,
		},
		{
			l:        XY,
			cs:       [][]Coord{},
			expected: nil,
		},
		{
			l:        XY,
			cs:       [][]Coord{{{1, 2}, {}}},
			expected: ErrStrideMismatch{Got: 0, Want: 2},
		},
		{
			l:        XY,
			cs:       [][]Coord{{{1, 2}, {1}}},
			expected: ErrStrideMismatch{Got: 1, Want: 2},
		},
		{
			l:        XY,
			cs:       [][]Coord{{{1, 2}, {3, 4}}},
			expected: nil,
		},
		{
			l:        XY,
			cs:       [][]Coord{{{1, 2}, {3, 4, 5}}},
			expected: ErrStrideMismatch{Got: 3, Want: 2},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := NewPolygon(tc.l).SetCoords(tc.cs)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPolygonSetSRID(t *testing.T) {
	assert.Equal(t, 4326, NewPolygon(NoLayout).SetSRID(4326).SRID())
	assert.Equal(t, 4326, Must(SetSRID(NewPolygon(NoLayout), 4326)).SRID())
}
