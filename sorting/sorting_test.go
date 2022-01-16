package sorting_test

import (
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/sorting"
)

func TestCompare2D(t *testing.T) {
	for i, tc := range []struct {
		c1, c2 []float64
		result bool
	}{
		{
			c1:     []float64{0, 0},
			c2:     []float64{0, 0},
			result: false,
		},
		{
			c1:     []float64{1, 0},
			c2:     []float64{0, 1},
			result: false,
		},
		{
			c1:     []float64{1, 0},
			c2:     []float64{0, 0},
			result: false,
		},
		{
			c1:     []float64{0, 1},
			c2:     []float64{0, 0},
			result: false,
		},
		{
			c1:     []float64{0, 0},
			c2:     []float64{0, 1},
			result: true,
		},
		{
			c1:     []float64{0, 0},
			c2:     []float64{1, 0},
			result: true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			require.Equal(t, tc.result, sorting.IsLess2D(tc.c1, tc.c2))
		})
	}
}

func TestNewFlatCoordSorting2D(t *testing.T) {
	for i, tc := range []struct {
		c1, result []float64
		Lay     geom.Layout
	}{
		{
			c1:     []float64{},
			result: []float64{},
			Lay: geom.XY,
		},
		{
			c1:     []float64{0, 0, 1, 1},
			result: []float64{0, 0, 1, 1},
			Lay: geom.XY,
		},
		{
			c1:     []float64{1, 0, 0, 1, 2, 2, 2, -2, -1, 0, 0, 0, 0, 0},
			result: []float64{-1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 2, -2, 2, 2},
			Lay: geom.XY,
		},
		{
			c1:     []float64{1, 0, 6, 0, 1, 6, 2, 2, 6, 2, -2, 6, -1, 0, 8, 0, 0, 6, 0, 0, 6},
			result: []float64{-1, 0, 8, 0, 0, 6, 0, 0, 6, 0, 1, 6, 1, 0, 6, 2, -2, 6, 2, 2, 6},
			Lay: geom.XYM,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual := make([]float64, len(tc.c1))
			copy(actual, tc.c1)
			sort.Sort(sorting.NewFlatCoordSorting2D(tc.Lay, actual))

			require.Equal(t, tc.result, actual)
		})
	}
}
