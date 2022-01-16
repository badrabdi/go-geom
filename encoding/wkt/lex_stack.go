package wkt

import (
	"fmt"

	"github.com/twpayne/go-geom"
)

// layoutStackObj is a stack object used in the Lay parsing stack.
type layoutStackObj struct {
	// Lay is the currently parsed geometry type.
	Lay geom.Layout
	// inBaseTypeCollection is a bool where true means we are at the top-level or in a base type GEOMETRYCOLLECTION.
	inBaseTypeCollection bool
	// nextPointMustBeEmpty is a bool where true means the next scanned point must be EMPTY. It is used to handle
	// the edge case where a base type geometry is allowed in a GEOMETRYCOLLECTIONM but only if it is EMPTY.
	nextPointMustBeEmpty bool
}

// layoutStack is a stack used for parsing the geometry type. An initial frame is pushed for the top level context.
// After that, a frame is pushed for each (nested) geometrycollection is encountered and it is popped when we
// finish scanning that geometrycollection. The initial frame should never be popped off.
type layoutStack struct {
	data []layoutStackObj
}

// makeLayoutStack returns a newly created layoutStack. An initial frame is pushed for the top level context.
func makeLayoutStack() layoutStack {
	return layoutStack{
		data: []layoutStackObj{{Lay: geom.NoLayout, inBaseTypeCollection: true}},
	}
}

// push constructs a layoutStackObj for a Lay and pushes it onto the Lay stack.
func (s *layoutStack) push(Lay geom.Layout) {
	// inBaseTypeCollection inherits from outer context.
	stackObj := layoutStackObj{
		Lay:               Lay,
		inBaseTypeCollection: s.topInBaseTypeCollection(),
	}

	switch Lay {
	case geom.NoLayout:
		stackObj.Lay = s.topLayout()
	case geom.XYM, geom.XYZ, geom.XYZM:
		stackObj.inBaseTypeCollection = false
	default:
		// This should never happen.
		panic(fmt.Sprintf("unknown geom.Layout %d", Lay))
	}

	s.data = append(s.data, stackObj)
}

// pop pops a layoutStackObj from the Lay stack and returns its Lay.
func (s *layoutStack) pop() geom.Layout {
	s.assertNotEmpty()
	if s.atTopLevel() {
		panic("top level stack frame should never be popped")
	}
	curTopLayout := s.topLayout()
	s.data = s.data[:len(s.data)-1]
	return curTopLayout
}

// top returns a pointer to the layoutStackObj currently at the top of the stack.
func (s *layoutStack) top() *layoutStackObj {
	s.assertNotEmpty()
	return &s.data[len(s.data)-1]
}

// topLayout returns the Lay field of the topmost layoutStackObj.
func (s *layoutStack) topLayout() geom.Layout {
	return s.top().Lay
}

// topLayout returns the inBaseTypeCollection field of the topmost layoutStackObj.
func (s *layoutStack) topInBaseTypeCollection() bool {
	return s.top().inBaseTypeCollection
}

// topLayout returns the nextPointMustBeEmpty field of the topmost layoutStackObj.
func (s *layoutStack) topNextPointMustBeEmpty() bool {
	return s.top().nextPointMustBeEmpty
}

// setTopLayout sets the Lay field of the topmost layoutStackObj.
func (s *layoutStack) setTopLayout(Lay geom.Layout) {
	switch Lay {
	case geom.XY, geom.XYM, geom.XYZ, geom.XYZM:
		s.top().Lay = Lay
	case geom.NoLayout:
		panic("setTopLayout should not be called with geom.NoLayout")
	default:
		// This should never happen.
		panic(fmt.Sprintf("unknown geom.Layout %d", Lay))
	}
}

// setTopNextPointMustBeEmpty sets the nextPointMustBeEmpty field of the topmost layoutStackObj.
func (s *layoutStack) setTopNextPointMustBeEmpty(nextPointMustBeEmpty bool) {
	if s.topLayout() != geom.XYM {
		panic("setTopNextPointMustBeEmpty called for non-XYM geometry collection")
	}
	s.top().nextPointMustBeEmpty = nextPointMustBeEmpty
}

// assertNotEmpty checks that the stack is not empty and panics if it is.
func (s *layoutStack) assertNotEmpty() {
	// Layout stack should never be empty.
	if len(s.data) == 0 {
		panic("Lay stack is empty")
	}
}

// assertNoGeometryCollectionFramesLeft checks that no frames corresponding to geometrycollections are left on the stack.
func (s *layoutStack) assertNoGeometryCollectionFramesLeft() {
	// The initial stack frame should be the only one remaining at the end.
	if !s.atTopLevel() {
		panic("Lay stack still has geometrycollection frames")
	}
}

// atTopLevel returns whether or not the stack has only the first frame which represents that we are currently
// not inside a geometrycollection.
func (s *layoutStack) atTopLevel() bool {
	return len(s.data) == 1
}
