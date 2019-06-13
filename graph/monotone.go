package graph

import (
	"github.com/pkg/errors"
	"github.com/simoncochrane/geoz/coord"
)

type MonotoneChain struct {
	Edge  *MonotoneChainEdge
	Index int
}

type MonotoneChainEdge struct {
	Edge   *Edge
	Points coord.Coordinates

	// the list of start/end indexes of the monotone chains.
	// includes the end point of the edge as a sentinel.
	StartIndexes []int
}

func NewMonotoneChainEdge(edge *Edge) (*MonotoneChainEdge, error) {
	startIndexes, err := chainStartIndexes(edge.Coordinates)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chain start indices for MonotoneChainEdge")
	}

	return &MonotoneChainEdge{
		Edge:         edge,
		Points:       edge.Coordinates,
		StartIndexes: startIndexes,
	}, nil
}

func chainStartIndexes(pts coord.Coordinates) ([]int, error) {
	start := 0
	startIndices := []int{start}
	for start < len(pts)-1 {
		last, err := findChainEnd(pts, start)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find chain end")
		}
		startIndices = append(startIndices, last)
		start = last
	}
	return startIndices, nil
}

func findChainEnd(pts coord.Coordinates, start int) (int, error) {
	chainQuad, err := QuadrantCoord(pts[start], pts[start+1])
	if err != nil {
		return 0, errors.Wrap(err, "failed to get Quadrant for coord")
	}
	last := start + 1
	for last < len(pts) {
		quad, err := QuadrantCoord(pts[last-1], pts[last])
		if err != nil {
			return 0, errors.Wrap(err, "failed to get Quadrant for coord")
		}
		if quad != chainQuad {
			break
		}
	}
	return last - 1, nil
}

func (mce *MonotoneChainEdge) MonotoneChain(chainIndex int) *MonotoneChain {
	return &MonotoneChain{
		Edge:  mce,
		Index: chainIndex,
	}
}

func (mce *MonotoneChainEdge) MinX(chainIndex int) float64 {
	x1 := mce.Points[mce.StartIndexes[chainIndex]].X
	x2 := mce.Points[mce.StartIndexes[chainIndex+1]].X
	if x1 < x2 {
		return x1
	}
	return x2
}

func (mce *MonotoneChainEdge) MaxX(chainIndex int) float64 {
	x1 := mce.Points[mce.StartIndexes[chainIndex]].X
	x2 := mce.Points[mce.StartIndexes[chainIndex+1]].X
	if x1 > x2 {
		return x1
	}
	return x2
}

func (mce *MonotoneChainEdge) Overlaps(start0, end0 int, other *MonotoneChainEdge, start1, end1 int) bool {
	return coord.EnvelopeIntersects(mce.Points[start0], mce.Points[end0], other.Points[start1], other.Points[end1])
}

func (mce *MonotoneChainEdge) ComputeIntersections(chainIndex0, chainIndex1 int, other *MonotoneChainEdge, si *SegmentIntersector) {
	computeIntersectionsForChain(
		mce, mce.StartIndexes[chainIndex0], mce.StartIndexes[chainIndex0+1],
		other, other.StartIndexes[chainIndex1], other.StartIndexes[chainIndex1+1],
		si)
}

func computeIntersectionsForChain(mce0 *MonotoneChainEdge, start0, end0 int, mce1 *MonotoneChainEdge, start1, end1 int,
	si *SegmentIntersector) {

	// terminating condition for the recursion
	if end0-start0 == 1 && end1-start1 == 1 {
		si.AddIntersections(mce0.Edge, start0, mce1.Edge, start1)
		return
	}

	// nothing to do if the envelopes of these chains don't overlap
	if !mce0.Overlaps(start0, end0, mce1, start1, end1) {
		return
	}

	// the chains overlap, so split each in half and iterate  (binary search)
	mid0 := (start0 + end0) / 2
	mid1 := (start1 + end1) / 2

	// Assert: mid != start or end (since we checked above for end - start <= 1)
	// check terminating conditions before recursing
	if start0 < mid0 {
		if start1 < mid1 {
			computeIntersectionsForChain(mce0, start0, mid0, mce1, start1, mid1, si)
		}
		if mid1 < end1 {
			computeIntersectionsForChain(mce0, start0, mid0, mce1, mid1, end1, si)
		}
	}
	if mid0 < end0 {
		if start1 < mid1 {
			computeIntersectionsForChain(mce0, mid0, end0, mce1, start1, mid1, si)
		}
		if mid1 < end1 {
			computeIntersectionsForChain(mce0, mid0, end0, mce1, mid1, end1, si)
		}
	}
}

func (mc *MonotoneChain) ComputeIntersections(other *MonotoneChain, si *SegmentIntersector) {
	mc.Edge.ComputeIntersections(mc.Index, other.Index, other.Edge, si)
}
