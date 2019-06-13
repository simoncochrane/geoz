package graph

import (
	"github.com/pkg/errors"
	"github.com/simoncochrane/geoz/coord"
	"github.com/simoncochrane/geoz/geom"
)

type Position int

const (
	PositionOn    Position = 0
	PositionLeft           = 1
	PositionRight          = 2
)

// TopologyLocation contains locations for a parent that is
// [0] On:    on the edge
// [1] Left:  left-hand side of the edge
// [2] Right: right-hand side of the edge
type TopologyLocation [3]coord.Location

func NewTopologyLocation() TopologyLocation {
	return TopologyLocation{coord.LocationNone, coord.LocationNone, coord.LocationNone}
}

func (tl TopologyLocation) IsNil() bool {
	for _, l := range tl {
		if l != coord.LocationNone {
			return false
		}
	}
	return true
}

// Label describes the relationship of a component of a topological graph to
// its neighbours.
type Label [2]TopologyLocation

func NewLabel() *Label {
	return &Label{
		NewTopologyLocation(),
		NewTopologyLocation(),
	}
}

// LocationAt returns the location for the label at the given index and position.
func (l Label) LocationAt(index int, pos Position) coord.Location {
	return l[index][pos]
}

func (l *Label) SetLocation(index int, loc coord.Location) {
	l[index][PositionOn] = loc
}

func (l *Label) SetLocations(index int, on, left, right coord.Location) {
	l[index][PositionOn] = on
	l[index][PositionLeft] = left
	l[index][PositionRight] = right
}

func (l *Label) IsNil(index int) bool {
	return l[index].IsNil()
}

func (l *Label) GeometryCount() int {
	count := 0
	if !l[0].IsNil() {
		count++
	}
	if !l[1].IsNil() {
		count++
	}
	return count
}

// Node represents a point in a graph.
type Node struct {
	Point coord.Coordinate
	Edges *EdgeEndStar
	Label *Label
}

func NewNode(point coord.Coordinate, edges *EdgeEndStar) *Node {
	return &Node{
		Point: point,
		Label: NewLabel(),
		Edges: edges,
	}
}

func (n *Node) SetLabelBoundary(index int) {
	loc := n.Label.LocationAt(index, PositionOn)

	newLoc := coord.LocationNone
	switch loc {
	case coord.LocationBoundary:
		newLoc = coord.LocationInterior
	case coord.LocationInterior:
		newLoc = coord.LocationBoundary
	default:
		newLoc = coord.LocationBoundary
	}
	n.Label.SetLocation(index, newLoc)
}

func (n *Node) Isolated() bool {
	return n.Label.GeometryCount() == 1
}

// Graph represents a topology graph, for use in calculating an intersection matrix.
// Note: currently only supports boundary node rule of "OGC_SFS_BOUNDARY_RULE"
type Graph struct {
	nodes map[coord.Coordinate]*Node

	// cached copy of the generated boundary nodes
	boundaryNodes []*Node

	edges       []*Edge
	lineEdgeMap *coord.CoordinatesMap

	UseBoundaryDeterminationRule bool

	boundaryNodeRule BoundaryNodeRule

	// the index of this geometry as an argument to a spatial function (used for labelling)
	argIndex int

	geometry *geom.Geometry
}

func NewGraph(parent *geom.Geometry, index int, useBoundaryDeterminationRule bool) (*Graph, error) {
	g := &Graph{
		geometry:         parent,
		argIndex:         index,
		lineEdgeMap:      coord.NewCoordinatesMap(),
		boundaryNodeRule: NewMod2BoundaryNodeRule(),

		UseBoundaryDeterminationRule: useBoundaryDeterminationRule,
	}
	if err := g.add(parent, index); err != nil {
		return nil, errors.Wrap(err, "failed to add geometry to graph")
	}
	return g, nil
}

func (gr *Graph) add(g *geom.Geometry, index int) error {
	switch g.Type {
	case geom.TypePoint:
		return gr.addPoint(index, g.Coord)
	case geom.TypeMultiPoint:
		return gr.addCollection(g, index)
	case geom.TypeLineString:
		return gr.addLineString(index, g.Line)
	case geom.TypeMultiLineString:
		return gr.addCollection(g, index)
	case geom.TypePolygon:
		return gr.addPolygon(index, g.Line, g.MultiLine)
	case geom.TypeMultiPolygon:
		gr.UseBoundaryDeterminationRule = false
		return gr.addCollection(g, index)
	case geom.TypeCollection:
		return gr.addCollection(g, index)
	}
	return errors.Errorf("unsupported geometry type for Graph: %v", g.Type)
}

func (gr *Graph) addCollection(g *geom.Geometry, index int) error {
	numGeoms := g.NumGeometries()
	for i := 0; i < numGeoms; i++ {
		if err := gr.add(g.GeometryN(i), index); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (gr *Graph) addPoint(index int, point coord.Coordinate) error {
	gr.insertPoint(index, point, coord.LocationInterior)
	return nil
}

func (gr *Graph) addLineString(index int, line coord.Coordinates) error {
	// TODO:
	//     Coordinate[] coord = CoordinateArrays.removeRepeatedPoints(line.getCoordinates());

	if len(line) < 2 {
		// TODO: deal with this case properly?
		return errors.Errorf("LineString has too few points for graph operation")
	}

	e := NewEdge(line)
	e.Label = LabelAt(index, On(coord.LocationInterior))

	gr.edges = append(gr.edges, e)
	gr.lineEdgeMap.Add(line, e)

	gr.insertBoundaryPoint(index, line[0])
	gr.insertBoundaryPoint(index, line[len(line)-1])

	return nil
}

func (gr *Graph) addPolygon(index int, shell coord.Coordinates, holes coord.MultiLine) error {
	if err := gr.insertPolygonRing(index, shell, coord.LocationExterior, coord.LocationInterior); err != nil {
		return errors.WithStack(err)
	}

	for _, hole := range holes {
		if err := gr.insertPolygonRing(index, hole, coord.LocationInterior, coord.LocationExterior); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func LabelAt(index int, location TopologyLocation) *Label {
	label := NewLabel()
	label[index] = location
	return label
}

func On(location coord.Location) TopologyLocation {
	tl := NewTopologyLocation()
	tl[0] = location
	return tl
}

func (gr *Graph) BoundaryNodes() []*Node {
	if gr.boundaryNodes != nil {
		return gr.boundaryNodes
	}

	var boundaryNodes []*Node
	for _, node := range gr.nodes {
		if node.Label.LocationAt(gr.argIndex, PositionOn) == coord.LocationBoundary {
			boundaryNodes = append(boundaryNodes, node)
		}
	}

	gr.boundaryNodes = boundaryNodes
	return boundaryNodes
}

func (gr *Graph) determineBoundary(boundaryCount int) coord.Location {
	if gr.boundaryNodeRule.InBoundary(boundaryCount) {
		return coord.LocationBoundary
	}
	return coord.LocationInterior
}

func (gr *Graph) addNode(point coord.Coordinate) *Node {
	node, has := gr.nodes[point]
	if !has {
		node = NewNode(point, nil)
		gr.nodes[point] = node
	}
	return node
}

func (gr *Graph) insertPoint(index int, point coord.Coordinate, location coord.Location) {
	node := gr.addNode(point)
	node.Label.SetLocation(index, location)
}

func (gr *Graph) insertBoundaryPoint(index int, point coord.Coordinate) {
	node := gr.addNode(point)

	boundaryCount := 1

	loc := node.Label.LocationAt(index, PositionOn)
	if loc == coord.LocationBoundary {
		boundaryCount++
	}

	newLoc := gr.determineBoundary(boundaryCount)
	node.Label.SetLocation(index, newLoc)
}

func (gr *Graph) insertPolygonRing(index int, points coord.Coordinates, cwLeft, cwRight coord.Location) error {
	if len(points) == 0 {
		return nil
	}

	// TODO: ??
	// Coordinate[] coord = CoordinateArrays.removeRepeatedPoints(lr.getCoordinates());

	if len(points) < 4 {
		return errors.Errorf("Polygon ring has too few points - found %d points", len(points))
	}

	left, right := cwLeft, cwRight
	if ccw, err := coord.IsCCW(points); err != nil {
		return errors.Wrapf(err, "failed to determine IsCCW for polygon ring %v", points)
	} else if ccw {
		left, right = cwRight, cwLeft
	}

	e := NewEdge(points)
	e.Label.SetLocations(index, coord.LocationBoundary, left, right)

	gr.edges = append(gr.edges, e)
	gr.lineEdgeMap.Add(points, e)

	gr.insertPoint(index, points[0], coord.LocationBoundary)

	return nil
}

// NOTE: computeAllSegments is only valid for ring geometries.
func (gr *Graph) computeSelfNodes(li coord.LineIntersector, computeAllSegments, isDoneIfProperInt bool) *SegmentIntersector {
	segmentIntersector := NewSegmentIntersector(li, true, false, isDoneIfProperInt)
	edgeSetIntersector := NewSimpleMCSweepLineIntersector()

	edgeSetIntersector.ComputeSelfIntersections(gr.edges, segmentIntersector, computeAllSegments)

	gr.addSelfIntersectionNodes(gr.argIndex)
	return segmentIntersector
}

func (gr *Graph) addSelfIntersectionNodes(argIndex int) {
	for _, edge := range gr.edges {
		eLoc := edge.Label.LocationAt(argIndex, PositionOn)
		for _, ei := range edge.eiList.nodeMap {
			gr.addSelfIntersectionNode(argIndex, ei.Coordinate, eLoc)
		}
	}
}

func (gr *Graph) addSelfIntersectionNode(argIndex int, point coord.Coordinate, location coord.Location) {
	if gr.IsBoundaryNode(argIndex, point) {
		return
	}

	if location == coord.LocationBoundary && gr.UseBoundaryDeterminationRule {
		gr.insertBoundaryPoint(argIndex, point)
	} else {
		gr.insertPoint(argIndex, point, location)
	}
}

func (gr *Graph) IsBoundaryNode(argIndex int, point coord.Coordinate) bool {
	node, has := gr.nodes[point]
	if !has {
		return false
	}

	if node.Label.LocationAt(argIndex, PositionOn) == coord.LocationBoundary {
		return true
	}

	return false
}

func (gr *Graph) computeEdgeIntersections(other *Graph, li coord.LineIntersector, includeProper bool) *SegmentIntersector {
	si := NewSegmentIntersector(li, includeProper, true, false)
	si.SetBoundaryNodes(gr.BoundaryNodes(), other.BoundaryNodes())

	esi := NewSimpleMCSweepLineIntersector()
	esi.ComputeEdgeIntersections(gr.edges, other.edges, si)

	return si
}
