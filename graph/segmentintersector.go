package graph

import "github.com/simoncochrane/geoz/coord"

type SegmentIntersector struct {
	lineIntersector coord.LineIntersector

	includeProper       bool
	recordIsolated      bool
	isDoneWhenProperInt bool

	hasIntersection   bool
	numIntersections  int
	hasProper         bool
	hasProperInterior bool

	properIntersectionPoint coord.Coordinate

	isDone bool

	boundaryNodes [2][]*Node
}

func NewSegmentIntersector(lineIntersector coord.LineIntersector,
	includeProper, recordIsolated, isDoneWhenProperInt bool) *SegmentIntersector {

	return &SegmentIntersector{
		lineIntersector:     lineIntersector,
		includeProper:       includeProper,
		recordIsolated:      recordIsolated,
		isDoneWhenProperInt: isDoneWhenProperInt,
	}
}

func (si *SegmentIntersector) SetBoundaryNodes(boundaryNodes0, boundaryNodes1 []*Node) {
	si.boundaryNodes[0] = boundaryNodes0
	si.boundaryNodes[1] = boundaryNodes1
}

func (si *SegmentIntersector) Done() bool {
	return si.isDone
}

func (si *SegmentIntersector) AddIntersections(e0 *Edge, segIndex0 int, e1 *Edge, segIndex1 int) {
	if e0 == e1 && segIndex0 == segIndex1 {
		return
	}

	p00 := e0.Coordinates[segIndex0]
	p01 := e0.Coordinates[segIndex0+1]
	p10 := e1.Coordinates[segIndex1]
	p11 := e1.Coordinates[segIndex1+1]

	si.lineIntersector.ComputeLineIntersection(p00, p01, p10, p11)

	// always record any non-proper intersections.
	// If includeProper is true, record any proper intersections as well.

	if si.lineIntersector.HasIntersection() {
		if si.recordIsolated {
			e0.Isolated = false
			e1.Isolated = false
		}
		si.numIntersections++

		// if the segments are adjacent they have at least one trivial intersection,
		// the shared endpoint. Don't bother adding it if it is the only intersection.
		if !si.isTrivialIntersection(e0, segIndex0, e1, segIndex1) {
			si.hasIntersection = true
			if si.includeProper || !si.lineIntersector.IsProper() {
				e0.AddIntersections(si.lineIntersector, segIndex0, 0)
				e1.AddIntersections(si.lineIntersector, segIndex1, 1)
			}
			if si.lineIntersector.IsProper() {
				si.properIntersectionPoint = si.lineIntersector.IntersectionAt(0)
				si.hasProper = true
				if si.isDoneWhenProperInt {
					si.isDone = true
				}
				if !isBoundaryPoint(si.lineIntersector, si.boundaryNodes) {
					si.hasProperInterior = true
				}
			}
		}
	}
}

func (si *SegmentIntersector) isTrivialIntersection(e0 *Edge, segIndex0 int, e1 *Edge, segIndex1 int) bool {
	if e0 == e1 {
		if si.lineIntersector.NumIntersections() == 1 {
			if adjacentSegments(segIndex0, segIndex1) {
				return true
			}
			if e0.Closed() {
				maxSegIndex := len(e0.Coordinates) - 1
				if (segIndex0 == 0 && segIndex1 == maxSegIndex) || (segIndex1 == 0 && segIndex0 == maxSegIndex) {
					return true
				}
			}
		}
	}
	return false
}

func adjacentSegments(i1, i2 int) bool {
	if i1 < 0 {
		i1 *= -1
	}
	if i2 < 0 {
		i2 *= -1
	}
	return i1 == i2
}

func isBoundaryPoint(li coord.LineIntersector, boundaryNodes [2][]*Node) bool {
	if boundaryNodes[0] == nil {
		return false
	}
	if isBoundaryPointInternal(li, boundaryNodes[0]) {
		return true
	}
	if isBoundaryPointInternal(li, boundaryNodes[1]) {
		return true
	}
	return false
}

func isBoundaryPointInternal(li coord.LineIntersector, boundaryNodes []*Node) bool {
	for _, node := range boundaryNodes {
		if li.IntersectsPoint(node.Point) {
			return true
		}
	}
	return false
}
