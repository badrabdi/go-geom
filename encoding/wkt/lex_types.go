package wkt

type geomFlatCoordsRepr struct {
	FlatCoord []float64
	ends      []int
}

func makeGeomFlatCoordsRepr(FlatCoord []float64) geomFlatCoordsRepr {
	return geomFlatCoordsRepr{FlatCoord: FlatCoord, ends: []int{len(FlatCoord)}}
}

func appendGeomFlatCoordsReprs(p1 geomFlatCoordsRepr, p2 geomFlatCoordsRepr) geomFlatCoordsRepr {
	if len(p1.ends) > 0 {
		p1LastEnd := p1.ends[len(p1.ends)-1]
		for i := range p2.ends {
			p2.ends[i] += p1LastEnd
		}
	}
	return geomFlatCoordsRepr{FlatCoord: append(p1.FlatCoord, p2.FlatCoord...), ends: append(p1.ends, p2.ends...)}
}

type multiPolygonFlatCoordsRepr struct {
	FlatCoord []float64
	endss     [][]int
}

func makeMultiPolygonFlatCoordsRepr(p geomFlatCoordsRepr) multiPolygonFlatCoordsRepr {
	if p.FlatCoord == nil {
		return multiPolygonFlatCoordsRepr{FlatCoord: nil, endss: [][]int{nil}}
	}
	return multiPolygonFlatCoordsRepr{FlatCoord: p.FlatCoord, endss: [][]int{p.ends}}
}

func appendMultiPolygonFlatCoordsRepr(
	p1 multiPolygonFlatCoordsRepr, p2 multiPolygonFlatCoordsRepr,
) multiPolygonFlatCoordsRepr {
	p1LastEndsLastEnd := 0
	for i := len(p1.endss) - 1; i >= 0; i-- {
		if len(p1.endss[i]) > 0 {
			p1LastEndsLastEnd = p1.endss[i][len(p1.endss[i])-1]
			break
		}
	}
	if p1LastEndsLastEnd > 0 {
		for i := range p2.endss {
			for j := range p2.endss[i] {
				p2.endss[i][j] += p1LastEndsLastEnd
			}
		}
	}
	return multiPolygonFlatCoordsRepr{
		FlatCoord: append(p1.FlatCoord, p2.FlatCoord...), endss: append(p1.endss, p2.endss...),
	}
}
