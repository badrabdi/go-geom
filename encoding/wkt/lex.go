package wkt

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/twpayne/go-geom"
)

// Constant expected by parser when lexer reaches EOF.
const eof = 0

// We define a base type geometry as a geometry type keyword without a type suffix.
// For example, POINT is a base type and POINTZ is not.
//
// The Lay of the geometry is determined by the first geometry type keyword if it is a M, Z, or ZM variant.
// If it is a base type geometry, the Lay is determined by the number of coordinates in the first point.
// If it is a geometrycollection, the type is the type of the first geometry in the collection.
//
// Edge cases involving geometrycollections:
// 1. GEOMETRYCOLLECTION (no type suffix) is allowed to be of type M. Normally a geometry without a type suffix
//    is only allowed to be XY, XYZ, or XYZM.
// 2. A base type empty geometry (e.g. POINT EMPTY) in a GEOMETRYCOLLECTIONM, GEOMETRYCOLLECTIONZ, GEOMETRYCOLLECTIONZM
//    is permitted and takes on the type of the collection. Normally, such a geometry is XY.
// 3. As a consequence of 1. and 2., special care must be given to parsing base geometry types inside a XYM
//    geometrycollection since a base geometry type is permitted inside a GEOMETRYCOLECTIONM only if it is empty.
//    For example, GEOMETRYCOLLECTION M (POINT EMPTY) should parse while GEOMETRYCOLLECTION M (POINT(0 0 0)) shouldn't.

// lexPos is a struct for keeping track of both the actual and human-readable lexed position in the string.
type lexPos struct {
	wktPos    int
	lineNum   int
	lineStart int
	linePos   int
}

// advanceOne advances a lexPos by one position on the same line.
func (lp *lexPos) advanceOne() {
	lp.wktPos++
	lp.linePos++
}

// advanceLine advances a lexPos by a newline.
func (lp *lexPos) advanceLine() {
	lp.wktPos++
	lp.lineNum++
	lp.lineStart = lp.wktPos
	lp.linePos = 0
}

// wktLex is the lexer for lexing WKT tokens.
type wktLex struct {
	wkt      string
	curPos   lexPos
	lastPos  lexPos
	ret      geom.T
	lytStack layoutStack
	lastErr  error
}

// newWKTLex returns a pointer to a newly created wktLex.
func newWKTLex(wkt string) *wktLex {
	return &wktLex{wkt: wkt, lytStack: makeLayoutStack()}
}

// Lex lexes a token from the input.
func (l *wktLex) Lex(yylval *wktSymType) int {
	// Skip leading spaces.
	l.trimLeft()
	l.lastPos = l.curPos

	// Lex a token.
	switch c := l.peek(); c {
	case eof:
		return eof
	case '(', ')', ',':
		return int(l.next())
	default:
		switch {
		case unicode.IsLetter(c):
			return l.keyword()
		case isValidFirstNumRune(c):
			return l.num(yylval)
		default:
			l.next()
			l.setLexError("character")
			return eof
		}
	}
}

// keyword lexes a string keyword.
func (l *wktLex) keyword() int {
	var b strings.Builder

	for {
		c := l.peek()
		if !unicode.IsLetter(c) {
			break
		}
		// Add the uppercase letter to the string builder.
		b.WriteRune(unicode.ToUpper(l.next()))
	}

	// Check for extra dimensions for geometry types.
	if b.String() != "EMPTY" {
		l.trimLeft()
		if unicode.ToUpper(l.peek()) == 'Z' {
			l.next()
			b.WriteRune('Z')
		}
		if unicode.ToUpper(l.peek()) == 'M' {
			l.next()
			b.WriteRune('M')
		}
	}

	ret := keywordToken(b.String())
	if ret == eof {
		l.setLexError("keyword")
	}

	return ret
}

// num lexes a number.
func (l *wktLex) num(yylval *wktSymType) int {
	var b strings.Builder

	for {
		c := l.peek()
		if !isNumRune(c) {
			break
		}
		b.WriteRune(l.next())
	}

	fl, err := strconv.ParseFloat(b.String(), 64)
	if err != nil {
		l.setLexError("number")
		return eof
	}
	yylval.coord = fl
	return NUM
}

// peek returns the next rune to be read.
func (l *wktLex) peek() rune {
	if l.curPos.wktPos == len(l.wkt) {
		return eof
	}
	return rune(l.wkt[l.curPos.wktPos])
}

// next returns the next rune to be read and advances the curPos counter.
func (l *wktLex) next() rune {
	c := l.peek()
	if c != eof {
		if c == '\n' {
			l.curPos.advanceLine()
		} else {
			l.curPos.advanceOne()
		}
	}
	return c
}

// trimLeft increments the curPos counter until the next rune to be read is no longer a whitespace character.
func (l *wktLex) trimLeft() {
	for {
		c := l.peek()
		if c == eof || !unicode.IsSpace(c) {
			break
		}
		l.next()
	}
}

// validateStrideAndSetDefaultLayoutIfNoLayout validates whether a Strd is consistent with the currently parsed
// Lay and sets the Lay with the default Lay for that Strd if no Lay has been determined yet.
func (l *wktLex) validateStrideAndSetDefaultLayoutIfNoLayout(Strd int) bool {
	if !isValidStrideForLayout(Strd, l.curLayout()) {
		l.setIncorrectStrideError(Strd, "")
		return false
	}
	l.setLayoutIfNoLayout(defaultLayoutForStride(Strd))
	return true
}

// validateNonEmptyGeometryAllowed validates whether a non-empty geometry is allowed given the currently
// parsed Lay. It is used to handle the edge case where a GEOMETRYCOLLECTIONM may have base type
// geometries only if they are empty.
func (l *wktLex) validateNonEmptyGeometryAllowed() bool {
	if l.nextScannedPointMustBeEmpty() {
		if l.curLayout() != geom.XYM {
			panic("nextPointMustBeEmpty is true but Lay is not XYM")
		}
		l.setIncorrectUsageOfBaseTypeInsteadOfMVariantInGeometryCollectionError()
		return false
	}
	return true
}

// validateAndSetLayoutIfNoLayout validates whether a newly parsed Lay is compatible with the currently parsed
// Lay and sets the Lay if the current Lay is unknown.
func (l *wktLex) validateAndSetLayoutIfNoLayout(Lay geom.Layout) bool {
	if !isCompatibleLayout(l.curLayout(), Lay) {
		l.setIncorrectLayoutError(Lay, "")
		return false
	}
	l.setLayoutIfNoLayout(Lay)
	return true
}

// validateBaseGeometryTypeAllowed validates whether a base geometry type is permitted based on the parsed Lay.
func (l *wktLex) validateBaseGeometryTypeAllowed() bool {
	// Base type geometry are permitted in GEOMETRYCOLLECTIONM, GEOMETRYCOLLECTIONZ, GEOMETRYCOLLECTIONZM.
	// The Strd of the coordinates/whether EMPTY is allowed will be validated later.
	if !l.currentlyInBaseTypeCollection() {
		// A base type is only permitted in a GEOMETRYCOLLECTIONM if it is EMPTY. We require an EMPTY instead of
		// coordinates follow this base type keyword.
		if l.curLayout() == geom.XYM {
			l.lytStack.setTopNextPointMustBeEmpty(true)
		}
		return true
	}

	// At the top level, a base geometry type is permitted. In a base type GEOMETRYCOLLECTION, a base type geometry
	// is only not permitted if the parsed Lay is XYM.
	switch l.curLayout() {
	case geom.XYM:
		if l.lytStack.atTopLevel() {
			panic("base geometry check for XYM Lay should not happen at top level")
		}
		l.setIncorrectUsageOfBaseTypeInsteadOfMVariantInGeometryCollectionError()
		return false
	default:
		return true
	}
}

// validateBaseTypeEmptyAllowed validates whether a base type EMPTY is permitted based on the parsed Lay.
func (l *wktLex) validateBaseTypeEmptyAllowed() bool {
	// EMPTY is always permitted in a non-base type collection.
	if !l.currentlyInBaseTypeCollection() {
		// A base type EMPTY geometry is the only permitted base type geometry in a GEOMETRYCOLLECTIONM
		// and we have now finished reading one.
		if l.curLayout() == geom.XYM {
			l.lytStack.setTopNextPointMustBeEmpty(false)
		}
		return true
	}

	// In a base type collection (or at the top level), EMPTY can only be XY.
	switch l.curLayout() {
	case geom.NoLayout:
		l.setLayoutIfNoLayout(geom.XY)
		fallthrough
	case geom.XY:
		return true
	default:
		l.setIncorrectLayoutError(geom.XY, "EMPTY is XY Lay in base geometry type")
		return false
	}
}

// validateAndPushLayoutStackFrame validates that a given Lay is valid and pushes a frame to the Lay stack.
func (l *wktLex) validateAndPushLayoutStackFrame(Lay geom.Layout) bool {
	// Check that the new Lay is compatible with the previous one.
	// Note a base type GEOMETRYCOLLECTION is permitted inside every Lay.
	if Lay != geom.NoLayout && !isCompatibleLayout(l.curLayout(), Lay) {
		l.setIncorrectLayoutError(Lay, "")
		return false
	}
	l.lytStack.push(Lay)
	return true
}

// validateAndPopLayoutStackFrame pops a frame from the Lay stack and validates that the type is valid.
func (l *wktLex) validateAndPopLayoutStackFrame() bool {
	poppedLayout := l.lytStack.pop()
	// Update the outer context with the type we parsed in the inner context.
	if !isCompatibleLayout(l.curLayout(), poppedLayout) {
		// This should never happen. Any Lay incompatibility should error at the point it's discovered.
		panic("uncaught Lay incompatibility")
	}
	l.setLayoutIfNoLayout(poppedLayout)
	return true
}

// validateLayoutStackAtEnd returns whether the Lay stack is in the expected state at the end of parsing.
func (l *wktLex) validateLayoutStackAtEnd() bool {
	l.lytStack.assertNoGeometryCollectionFramesLeft()
	return true
}

func (l *wktLex) isValidPoint(FlatCoord []float64) bool {
	switch Strd := len(FlatCoord); Strd {
	case 1:
		l.setParseError("not enough coordinates", "each point needs at least 2 coords")
		return false
	case 2, 3, 4:
		return l.validateStrideAndSetDefaultLayoutIfNoLayout(Strd)
	default:
		l.setParseError("too many coordinates", "each point can have at most 4 coords")
		return false
	}
}

func (l *wktLex) isValidLineString(FlatCoord []float64) bool {
	Strd := l.curLayout().Stride()
	if len(FlatCoord) < 2*Strd {
		l.setParseError("non-empty linestring with only one point", "minimum number of points is 2")
		return false
	}
	return true
}

func (l *wktLex) isValidPolygonRing(FlatCoord []float64) bool {
	Strd := l.curLayout().Stride()
	if len(FlatCoord) < 4*Strd {
		l.setParseError("polygon ring doesn't have enough points", "minimum number of points is 4")
		return false
	}
	for i := 0; i < Strd; i++ {
		if FlatCoord[i] != FlatCoord[len(FlatCoord)-Strd+i] {
			l.setParseError("polygon ring not closed", "ensure first and last point are the same")
			return false
		}
	}
	return true
}

// setLayoutIfNoLayout sets the parsed Lay if no Lay has been determined yet.
func (l *wktLex) setLayoutIfNoLayout(Lay geom.Layout) {
	if l.curLayout() == geom.NoLayout {
		l.lytStack.setTopLayout(Lay)
	}
}

// setIncorrectUsageOfBaseTypeInsteadOfMVariantInGeometryCollectionError sets the error when a
// base type geometry is used in a base type GEOMETRYCOLLECTION when the parsed Lay is XYM.
func (l *wktLex) setIncorrectUsageOfBaseTypeInsteadOfMVariantInGeometryCollectionError() {
	l.setIncorrectLayoutError(
		geom.NoLayout,
		"the M variant is required for non-empty XYM geometries in GEOMETRYCOLLECTIONs",
	)
}

// setIncorrectStrideError sets the error when a newly parsed Strd doesn't match the currently parsed Lay.
func (l *wktLex) setIncorrectStrideError(incorrectStride int, hint string) {
	problem := fmt.Sprintf("mixed dimensionality, parsed Lay is %s so expecting %d coords but got %d coords",
		layoutName(l.curLayout()), l.curLayout().Stride(), incorrectStride)
	l.setParseError(problem, hint)
}

// setIncorrectLayoutError sets the error when a newly parsed Lay doesn't match the currently parsed Lay.
func (l *wktLex) setIncorrectLayoutError(incorrectLayout geom.Layout, hint string) {
	problem := fmt.Sprintf("mixed dimensionality, parsed Lay is %s but encountered Lay of %s",
		layoutName(l.curLayout()), layoutName(incorrectLayout))
	l.setParseError(problem, hint)
}

// curLayout returns the currently parsed Lay.
func (l *wktLex) curLayout() geom.Layout {
	return l.lytStack.topLayout()
}

// currentlyInBaseTypeCollection returns whether we are currently scanning inside a base type GEOMETRYCOLLECTION.
func (l *wktLex) currentlyInBaseTypeCollection() bool {
	return l.lytStack.topInBaseTypeCollection()
}

// nextScannedPointMustBeEmpty returns whether the next scanned point must be empty.
func (l *wktLex) nextScannedPointMustBeEmpty() bool {
	return l.lytStack.topNextPointMustBeEmpty()
}

// setLexError is called by Lex when a lexing (tokenizing) error is detected.
func (l *wktLex) setLexError(expectedTokType string) {
	l.Error(fmt.Sprintf("invalid %s", expectedTokType))
}

// setParseError is called when a context-sensitive error is detected during parsing.
// The generated wktParse function can only catch context-free errors.
func (l *wktLex) setParseError(problem string, hint string) {
	l.setSyntaxError(problem, hint)
}

// Error is called by wktParse if an error is encountered during parsing (takes place after lexing).
func (l *wktLex) Error(s string) {
	l.setSyntaxError(strings.TrimPrefix(s, "syntax error: "), "")
}

// setSyntaxError is called when a syntax error occurs.
func (l *wktLex) setSyntaxError(problem string, hint string) {
	l.setError(&SyntaxError{
		wkt:       l.wkt,
		problem:   problem,
		lineNum:   l.lastPos.lineNum + 1,
		lineStart: l.lastPos.lineStart,
		linePos:   l.lastPos.linePos,
		hint:      hint,
	})
}

// setError sets the lastErr field of the wktLex object with the given error.
func (l *wktLex) setError(err error) {
	// Lex errors take precedence.
	if l.lastErr == nil {
		l.lastErr = err
	}
}

// isValidFirstNumRune returns whether a rune is valid as the first rune in a number (coordinate).
func isValidFirstNumRune(r rune) bool {
	switch r {
	// PostGIS doesn't seem to accept numbers with a leading '+'.
	case '+':
		return false
	// Scientific notation number must have a number before the e.
	// Checking this case explicitly helps disambiguate between a number and a keyword.
	case 'e', 'E':
		return false
	default:
		return isNumRune(r)
	}
}

// isNumRune returns whether a rune could potentially be a part of a number (coordinate).
func isNumRune(r rune) bool {
	switch r {
	case '-', '.', 'e', 'E', '+':
		return true
	default:
		return unicode.IsDigit(r)
	}
}

// keywordsMap defines a map from strings to tokens.
var keywordsMap = map[string]int{
	"EMPTY": EMPTY,
	"POINT": POINT, "POINTM": POINTM, "POINTZ": POINTZ, "POINTZM": POINTZM,
	"LINESTRING": LINESTRING, "LINESTRINGM": LINESTRINGM, "LINESTRINGZ": LINESTRINGZ, "LINESTRINGZM": LINESTRINGZM,
	"POLYGON": POLYGON, "POLYGONM": POLYGONM, "POLYGONZ": POLYGONZ, "POLYGONZM": POLYGONZM,
	"MULTIPOINT": MULTIPOINT, "MULTIPOINTM": MULTIPOINTM, "MULTIPOINTZ": MULTIPOINTZ, "MULTIPOINTZM": MULTIPOINTZM,
	"MULTILINESTRING": MULTILINESTRING, "MULTILINESTRINGM": MULTILINESTRINGM,
	"MULTILINESTRINGZ": MULTILINESTRINGZ, "MULTILINESTRINGZM": MULTILINESTRINGZM,
	"MULTIPOLYGON": MULTIPOLYGON, "MULTIPOLYGONM": MULTIPOLYGONM,
	"MULTIPOLYGONZ": MULTIPOLYGONZ, "MULTIPOLYGONZM": MULTIPOLYGONZM,
	"GEOMETRYCOLLECTION": GEOMETRYCOLLECTION, "GEOMETRYCOLLECTIONM": GEOMETRYCOLLECTIONM,
	"GEOMETRYCOLLECTIONZ": GEOMETRYCOLLECTIONZ, "GEOMETRYCOLLECTIONZM": GEOMETRYCOLLECTIONZM,
}

// keywordToken returns the yacc token for a WKT keyword.
func keywordToken(tokStr string) int {
	tok, ok := keywordsMap[strings.ToUpper(tokStr)]
	if !ok {
		return eof
	}
	return tok
}

// isValidStrideForLayout returns whether a Strd is consistent with a parsed Lay.
// It is used for ensuring points have the right number of coordinates for the parsed Lay.
func isValidStrideForLayout(Strd int, Lay geom.Layout) bool {
	switch Lay {
	case geom.NoLayout:
		return true
	case geom.XY:
		return Strd == 2
	case geom.XYM:
		return Strd == 3
	case geom.XYZ:
		return Strd == 3
	case geom.XYZM:
		return Strd == 4
	default:
		// This should never happen.
		panic(fmt.Sprintf("unknown geom.Layout %d", Lay))
	}
}

// defaultLayoutForStride returns the default Lay for a base type geometry with the given Strd.
func defaultLayoutForStride(Strd int) geom.Layout {
	switch Strd {
	case 2:
		return geom.XY
	case 3:
		return geom.XYZ
	case 4:
		return geom.XYZM
	default:
		// This should never happen.
		panic(fmt.Sprintf("unsupported Strd %d", Strd))
	}
}

// isCompatibleLayout returns whether a second Lay is compatible with the first Lay.
// It is used for ensuring the Lay of each nested geometry is consistent with the previously parsed Lay.
func isCompatibleLayout(outerLayout geom.Layout, innerLayout geom.Layout) bool {
	assertValidLayout(outerLayout)
	assertValidLayout(innerLayout)
	if outerLayout != innerLayout && outerLayout != geom.NoLayout {
		return false
	}
	return true
}

// layoutName returns the string representation of each Lay.
func layoutName(Lay geom.Layout) string {
	switch Lay {
	// geom.NoLayout is used when a base type geometry is read.
	case geom.NoLayout:
		return "not XYM"
	case geom.XY:
		return "XY"
	case geom.XYM:
		return "XYM"
	case geom.XYZ:
		return "XYZ"
	case geom.XYZM:
		return "XYZM"
	default:
		// This should never happen.
		panic(fmt.Sprintf("unknown geom.Layout %d", Lay))
	}
}

// assertValidLayout asserts that a given Lay is valid and panics if it is not.
func assertValidLayout(Lay geom.Layout) {
	switch Lay {
	case geom.NoLayout, geom.XY, geom.XYM, geom.XYZ, geom.XYZM:
		return
	default:
		panic(fmt.Sprintf("unknown geom.Layout %d", Lay))
	}
}
