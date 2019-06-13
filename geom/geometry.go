package geom

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/simoncochrane/geoz/coord"
)

const (
	TypePoint Type = iota
	TypeMultiPoint
	TypeLineString
	TypeMultiLineString
	// TypeLinearRing
	TypePolygon
	TypeMultiPolygon
	TypeCollection
)

type Type int

func (t Type) String() string {
	switch t {
	case TypePoint:
		return "Point"
	case TypeMultiPoint:
		return "MultiPoint"
	case TypeLineString:
		return "LineString"
	case TypeMultiLineString:
		return "MultiLineString"
	case TypePolygon:
		return "Polygon"
	case TypeMultiPolygon:
		return "MultiPolygon"
	case TypeCollection:
		return "Collection"
	}
	return "Unknown"
}

type Coordinate = coord.Coordinate
type Coordinates = coord.Coordinates
type MultiLine = coord.MultiLine

type Geometry struct {
	Type       Type
	Coord      Coordinate
	Line       Coordinates
	MultiLine  MultiLine
	Collection []*Geometry

	envelope *coord.Envelope
}

func NewPoint(point Coordinate) (*Geometry, error) {
	return &Geometry{
		Type:  TypePoint,
		Coord: point,
	}, nil
}

func NewMultiPoint(points []*Geometry) (*Geometry, error) {
	for _, point := range points {
		if point.Type != TypePoint {
			return nil, errors.Errorf("MultiPoint may only contain geometry of type Point, found %v", point.Type)
		}
	}
	return &Geometry{
		Type:       TypeMultiPoint,
		Collection: points,
	}, nil
}

func NewLineString(line Coordinates) (*Geometry, error) {
	return &Geometry{
		Type: TypeLineString,
		Line: line,
	}, nil
}

func NewMultiLineString(lines []*Geometry) (*Geometry, error) {
	for _, line := range lines {
		if line.Type != TypeLineString {
			return nil, errors.Errorf("MultiLineString may only contain geometry of type LineString, found %v", line.Type)
		}
	}
	return &Geometry{
		Type:       TypeMultiLineString,
		Collection: lines,
	}, nil
}

func NewPolygon(shell Coordinates, interior MultiLine) (*Geometry, error) {
	// TODO: validate polygon points
	return &Geometry{
		Type:      TypePolygon,
		Line:      shell,
		MultiLine: interior,
	}, nil
}

func NewMultiPolygon(polygons []*Geometry) (*Geometry, error) {
	for _, poly := range polygons {
		if poly.Type != TypePolygon {
			return nil, errors.Errorf("MultiPolygon may only contain geometry of type Polygon, found %v", poly.Type)
		}
	}
	return &Geometry{
		Type:       TypeMultiPolygon,
		Collection: polygons,
	}, nil
}

func NewCollection(collection []*Geometry) (*Geometry, error) {
	return &Geometry{
		Type:       TypeCollection,
		Collection: collection,
	}, nil
}

// IsMulti returns true if the geometry is a multi or collection-based geometry.
func (g *Geometry) IsMulti() bool {
	switch g.Type {
	case TypeMultiPoint:
	case TypeMultiLineString:
	case TypeMultiPolygon:
	case TypeCollection:
		return true
	}
	return false
}

// IsRings returns true if the geometry type represents a type containing valid rings.
func (g *Geometry) IsRings() bool {
	switch g.Type {
	//case TypeLinearRing:
	case TypePolygon:
	case TypeMultiPolygon:
		return true
	}
	return false
}

func (g *Geometry) Dimension() int {
	switch g.Type {
	case TypePoint:
		return 0
	case TypeMultiPoint:
		return 0
	case TypeLineString:
		return 1
	case TypeMultiLineString:
		return 1
	case TypePolygon:
		return 2
	case TypeMultiPolygon:
	case TypeCollection:
		dim := -1
		for _, col := range g.Collection {
			if d := col.Dimension(); d > dim {
				dim = d
			}
		}
		return dim
	}

	panic(fmt.Sprintf("Unknown geometry type: %v", g.Type))
}

func (g *Geometry) BoundaryDimension() int {
	switch g.Type {
	case TypePoint:
		return -1
	case TypeMultiPoint:
		return -1
	case TypeLineString:
		if g.IsClosed() {
			return -1
		}
		return 0
	case TypeMultiLineString:
		if g.IsClosed() {
			return -1
		}
		return 0
	case TypePolygon:
		return 1
	case TypeMultiPolygon:
	case TypeCollection:
		dim := -1
		for _, col := range g.Collection {
			if d := col.BoundaryDimension(); d > dim {
				dim = d
			}
		}
		return dim
	}

	panic(fmt.Sprintf("Unknown geometry type: %v", g.Type))
}

func (g *Geometry) IsClosed() bool {
	if g.IsEmpty() {
		return false
	}

	switch g.Type {
	case TypeLineString:
		return g.Line[0].Equals2D(g.Line[len(g.Line)-1])
	case TypePolygon:
		return true
	}

	return false
}

func (g *Geometry) IsEmpty() bool {
	switch g.Type {
	case TypePoint:
		return false
	case TypeLineString:
	case TypePolygon:
		return len(g.Line) == 0
	case TypeMultiPoint:
	case TypeMultiLineString:
	case TypeMultiPolygon:
	case TypeCollection:
		for _, c := range g.Collection {
			if !c.IsEmpty() {
				return false
			}
		}
		return true
	}

	panic(fmt.Sprintf("Unknown geometry type: %v", g.Type))
}

func (g *Geometry) Envelope() *coord.Envelope {
	if g.envelope != nil {
		return g.envelope
	}

	var env *coord.Envelope

	if g.IsMulti() {
		for i, col := range g.Collection {
			if i == 0 {
				env = col.Envelope()
			} else {
				env.ExpandEnvelope(col.Envelope())
			}
		}
	}

	switch g.Type {
	case TypePoint:
		env = g.Coord.Envelope()
	case TypeLineString:
	case TypePolygon:
		env = g.Line.Envelope()
	}

	g.envelope = env
	return env
}

///////////////////////////////////////////////////////

func (g *Geometry) NumGeometries() int {
	return len(g.Collection)
}

func (g *Geometry) GeometryN(i int) *Geometry {
	return g.Collection[i]
}

//func (g *Geometry) Relate(other *Geometry) graph.IntersectionMatrix {
//	return nil
//	//return operation.NewRelate(g, other).IntersectionMatrix()
//}

//////////////////////////////////////////////////////////

// TODO: always use float for now.
//type PrecisionModel struct {
//}

//func (g *Geometry) PrecisionModel() *PrecisionModel {
//}
