package graph

import (
	"github.com/pkg/errors"
	"github.com/simoncochrane/geoz/coord"
)

type EdgeIntersection struct {
	// the point of intersection
	Coordinate coord.Coordinate

	// the index of the containing line segment in the parent edge
	SegmentIndex int

	// the edge distance of this point along the containing line segment
	Distance float64
}

type EdgeIntersectionList struct {
	// parent edge
	edge *Edge

	nodeMap map[EdgeIntersection]*EdgeIntersection
}

func NewEdgeIntersectionList(edge *Edge) *EdgeIntersectionList {
	return &EdgeIntersectionList{
		edge:    edge,
		nodeMap: map[EdgeIntersection]*EdgeIntersection{},
	}
}

func (eil *EdgeIntersectionList) Add(intPt coord.Coordinate, segmentIndex int, dist float64) *EdgeIntersection {
	ei := EdgeIntersection{
		Coordinate:   intPt,
		SegmentIndex: segmentIndex,
		Distance:     dist,
	}
	if ei, has := eil.nodeMap[ei]; has {
		return ei
	}
	eil.nodeMap[ei] = &ei
	return &ei
}

type EdgeEnd struct {
	Parent *Edge
}

type EdgeEndStar struct {
	// EdgeMap is a map that maintains the edges in sorted order around the node.
	EdgeMap map[EdgeEnd]interface{}

	// EdgeList is a list of all outgoing edges in the result, in CCW order.
	EdgeList []EdgeEnd

	pointInAreaLocation [2]coord.Location
}

func NewEdgeEndStar() *EdgeEndStar {
	return &EdgeEndStar{
		EdgeMap:             map[EdgeEnd]interface{}{},
		pointInAreaLocation: [2]coord.Location{coord.LocationNone, coord.LocationNone},
	}
}

type Edge struct {
	Coordinates coord.Coordinates
	Label       *Label

	Isolated bool

	monotoneChainEdge *MonotoneChainEdge
	eiList            *EdgeIntersectionList
}

func NewEdge(coords coord.Coordinates) *Edge {
	e := &Edge{
		Coordinates: coords,
		Label:       NewLabel(),
		Isolated:    true,
	}
	e.eiList = NewEdgeIntersectionList(e)
	return e
}

func (e *Edge) Closed() bool {
	return e.Coordinates[0].Equals(e.Coordinates[len(e.Coordinates)-1])
}

func (e *Edge) MonotoneChainEdge() (*MonotoneChainEdge, error) {
	if e.monotoneChainEdge == nil {
		mce, err := NewMonotoneChainEdge(e)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create MonotoneChainEdge for edge %v", e)
		}
		e.monotoneChainEdge = mce
	}
	return e.monotoneChainEdge, nil
}

func (e *Edge) AddIntersections(li coord.LineIntersector, segmentIndex, geomIndex int) {
	for i := 0; i < li.NumIntersections(); i++ {
		e.AddIntersection(li, segmentIndex, geomIndex, i)
	}
}

func (e *Edge) AddIntersection(li coord.LineIntersector, segmentIndex, geomIndex, intIndex int) {
	intPt := li.IntersectionAt(intIndex)

	normalizedSegmentIndex := segmentIndex
	dist := li.EdgeDistance(geomIndex, intIndex)

	// normalize the intersection point location
	nextSegIndex := normalizedSegmentIndex + 1
	if nextSegIndex < len(e.Coordinates) {
		nextPt := e.Coordinates[nextSegIndex]

		// Normalize segment index if intPt falls on vertex
		// The check for point equality is 2D only - Z values are ignored
		if intPt.Equals2D(nextPt) {
			normalizedSegmentIndex = nextSegIndex
			dist = 0.0
		}
	}

	// Add the intersection point to edge intersection list.
	e.eiList.Add(intPt, normalizedSegmentIndex, dist)
}
