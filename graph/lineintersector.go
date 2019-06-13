package graph

/*
type LineIntersectionResult int

const (
	LineIntersectionNone      LineIntersectionResult = 0
	LineIntersectionPoint                            = 1
	LineIntersectionCollinear                        = 2
)

type LineIntersector interface {
	HasIntersection() bool
	NumIntersections() int
	IsProper() bool

	IntersectionAt(intIndex int) coord.Coordinate
	EdgeDistance(segmentIndex, intIndex int) float64

	// IntersectsPoint tests whether a point is a intersection point of two line segments.
	IntersectsPoint(point coord.Coordinate) bool

	// ComputeIntersection computes the intersection of the lines p1-p2 and p3-p4.
	ComputeIntersection(p1, p2, p3, p4 coord.Coordinate)
}

type RobustLineIntersector struct {
	result   LineIntersectionResult
	isProper bool

	inputLines [2][2]coord.Coordinate
	intPts     [2]coord.Coordinate
}

func NewRobustLineIntersector() *RobustLineIntersector {
	return &RobustLineIntersector{}
}

func (rli *RobustLineIntersector) HasIntersection() bool {
	return rli.result != LineIntersectionNone
}

func (rli *RobustLineIntersector) NumIntersections() int {
	return int(rli.result)
}

func (rli *RobustLineIntersector) IsProper() bool {
	return rli.HasIntersection() && rli.isProper
}

func (rli *RobustLineIntersector) IntersectionAt(intIndex int) coord.Coordinate {
	return rli.intPts[intIndex]
}

func (rli *RobustLineIntersector) EdgeDistance(segmentIndex, intIndex int) float64 {
	return rli.computeEdgeDistance(rli.intPts[intIndex], rli.inputLines[segmentIndex][0], rli.inputLines[segmentIndex][1])
}

func (rli *RobustLineIntersector) computeEdgeDistance(p, p0, p1 coord.Coordinate) float64 {
	dx := math.Abs(p1.X - p0.X)
	dy := math.Abs(p1.Y - p0.Y)

	dist := -1.0 // sentinel value
	if p.Equals(p0) {
		dist = 0.0
	} else if p.Equals(p1) {
		if dx > dy {
			dist = dx
		} else {
			dist = dy
		}
	} else {
		pdx := math.Abs(p.X - p0.X)
		pdy := math.Abs(p.Y - p0.Y)
		if dx > dy {
			dist = pdx
		} else {
			dist = pdy
		}

		// FIXME: hack to ensure that non-endpoints always have a non-zero distance
		if dist == 0.0 && !p.Equals(p0) {
			dist = math.Max(pdx, pdy)
		}
	}

	// assert
	if dist == 0.0 && !p.Equals(p0) {
		panic("Bad distance calculation")
	}

	return dist
}

func (rli *RobustLineIntersector) IntersectsPoint(point coord.Coordinate) bool {
	for _, intPt := range rli.intPts {
		if intPt.Equals2D(point) {
			return true
		}

	}
	return false
}

func (rli *RobustLineIntersector) ComputeIntersection(p1, p2, q1, q2 coord.Coordinate) {
	rli.inputLines[0][0] = p1
	rli.inputLines[0][1] = p2
	rli.inputLines[1][0] = q1
	rli.inputLines[1][1] = q2

	rli.result = rli.computeIntersectionResult(p1, p2, q1, q2)
}

func (rli *RobustLineIntersector) computeIntersectionResult(p1, p2, q1, q2 coord.Coordinate) LineIntersectionResult {
	rli.isProper = false

	// first try a fast test to see if the envelopes of the lines intersect
	if !coord.EnvelopeIntersects(p1, p2, q1, q2) {
		return LineIntersectionNone
	}

	// for each endpoint, compute which side of the other segment it lies
	// if both endpoints lie on the same side of the other segment,
	// the segments do not intersect
	pq1 := coord.OrientationIndex(p1, p2, q1)
	pq2 := coord.OrientationIndex(p1, p2, q2)

	if (pq1 > 0 && pq2 > 0) || (pq1 < 0 && pq2 < 0) {
		return LineIntersectionNone
	}

	qp1 := coord.OrientationIndex(q1, q2, p1)
	qp2 := coord.OrientationIndex(q1, q2, p2)

	if (qp1 > 0 && qp2 > 0) || (qp1 < 0 && qp2 < 0) {
		return LineIntersectionNone
	}

	collinear := pq1 == 0 && pq2 == 0 && qp1 == 0 && qp2 == 0
	if collinear {
		return rli.computeCollinearIntersection(p1, p2, q1, q2)
	}

	// At this point we know that there is a single intersection point (since the lines are not collinear).

	//  Check if the intersection is an endpoint. If it is, copy the endpoint as
	//  the intersection point. Copying the point rather than computing it
	//  ensures the point has the exact value, which is important for
	//  robustness. It is sufficient to simply check for an endpoint which is on
	//  the other line, since at this point we know that the inputLines must
	//  intersect.

	if pq1 == 0 || pq2 == 0 || qp1 == 0 || qp2 == 0 {
		rli.isProper = false

		// Check for two equal endpoints.
		// This is done explicitly rather than by the orientation tests
		// below in order to improve robustness.
		//
		// [An example where the orientation tests fail to be consistent is
		// the following (where the true intersection is at the shared endpoint
		// POINT (19.850257749638203 46.29709338043669)
		//
		// LINESTRING ( 19.850257749638203 46.29709338043669, 20.31970698357233 46.76654261437082 )
		// and
		// LINESTRING ( -48.51001596420236 -22.063180333403878, 19.850257749638203 46.29709338043669 )
		//
		// which used to produce the INCORRECT result: (20.31970698357233, 46.76654261437082, NaN)

		if p1.Equals2D(q1) || p1.Equals2D(q2) {
			rli.intPts[0] = p1
		} else if p2.Equals2D(q1) || p2.Equals2D(q2) {
			rli.intPts[0] = p2
		} else {
			// Now check to see if any endpoint lies on the interior of the other segment.
			if pq1 == 0 {
				rli.intPts[0] = q1
			} else if pq2 == 0 {
				rli.intPts[0] = q2
			} else if qp1 == 0 {
				rli.intPts[0] = p1
			} else if qp2 == 0 {
				rli.intPts[0] = p2
			}
		}
	} else {
		rli.isProper = true
		intPt := rli.intersection(p1, p2, q1, q2)
		rli.intPts[0] = intPt
	}
	return LineIntersectionPoint
}

func (rli *RobustLineIntersector) computeCollinearIntersection(p1, p2, q1, q2 coord.Coordinate) LineIntersectionResult {
	p1q1p2 := coord.EnvelopeIntersectsPoint(p1, p2, q1)
	p1q2p2 := coord.EnvelopeIntersectsPoint(p1, p2, q2)
	q1p1q2 := coord.EnvelopeIntersectsPoint(q1, q2, p1)
	q1p2q2 := coord.EnvelopeIntersectsPoint(q1, q2, p2)

	if p1q1p2 && p1q2p2 {
		rli.intPts[0] = q1
		rli.intPts[1] = q2
		return LineIntersectionCollinear
	}
	if q1p1q2 && q1p2q2 {
		rli.intPts[0] = p1
		rli.intPts[1] = p2
		return LineIntersectionCollinear
	}
	if p1q1p2 && q1p1q2 {
		rli.intPts[0] = q1
		rli.intPts[1] = p1
		if q1.Equals(p1) && !p1q2p2 && !q1p2q2 {
			return LineIntersectionPoint
		}
		return LineIntersectionCollinear
	}
	if p1q1p2 && q1p2q2 {
		rli.intPts[0] = q1
		rli.intPts[1] = p2
		if q1.Equals(p2) && !p1q2p2 && !q1p1q2 {
			return LineIntersectionPoint
		}
		return LineIntersectionCollinear
	}
	if p1q2p2 && q1p1q2 {
		rli.intPts[0] = q2
		rli.intPts[1] = p1
		if q2.Equals(p1) && !p1q1p2 && !q1p2q2 {
			return LineIntersectionPoint
		}
		return LineIntersectionCollinear
	}
	if p1q2p2 && q1p2q2 {
		rli.intPts[0] = q2
		rli.intPts[1] = p2
		if q2.Equals(p2) && !p1q1p2 && !q1p1q2 {
			return LineIntersectionPoint
		}
		return LineIntersectionCollinear
	}
	return LineIntersectionNone
}

// intersection computes the actual value of the intersection point.
// To obtain the maximum precision from the intersection calculation,
// the coordinates are normalized by subtracting the minimum
// ordinate values (in absolute value). This has the effect of
// removing common significant digits from the calculation to
// maintain more bits of precision.
func (rli *RobustLineIntersector) intersection(p1, p2, q1, q2 coord.Coordinate) coord.Coordinate {
	intPt := rli.intersectionWithNormalization(p1, p2, q1, q2)

	if !rli.isInSegmentEnvelopes(intPt) {
		intPt = nearestEndpoint(p1, p2, q1, q2)
	}
	// TODO: use precision models?
	//if precisionModel != nil {
	//  precisionModel.makePrecise(intPt)
	//}
	return intPt
}

func (rli *RobustLineIntersector) intersectionWithNormalization(p1, p2, q1, q2 coord.Coordinate) coord.Coordinate {
	n1, n2, n3, n4, normPt := normalizeToEnvCentre(p1, p2, q1, q2)

	intPt := rli.safeHCoordinateIntersection(n1, n2, n3, n4)

	intPt.X += normPt.X
	intPt.Y += normPt.Y

	return intPt
}

func normalizeToEnvCentre(p1, p2, q1, q2 coord.Coordinate) (n00, n01, n10, n11, normPt coord.Coordinate) {
	minX0 := math.Min(p1.X, p2.X)
	minY0 := math.Min(p1.Y, p2.Y)
	maxX0 := math.Max(p1.X, p2.X)
	maxY0 := math.Max(p1.Y, p2.Y)

	minX1 := math.Min(q1.X, q2.X)
	minY1 := math.Min(q1.Y, q2.Y)
	maxX1 := math.Max(q1.X, q2.X)
	maxY1 := math.Max(q1.Y, q2.Y)

	intMinX := math.Max(minX0, minX1)
	intMaxX := math.Min(maxX0, maxX1)
	intMinY := math.Max(minY0, minY1)
	intMaxY := math.Min(maxY0, maxY1)

	intMidX := (intMinX + intMaxX) / 2.0
	intMidY := (intMinY + intMaxY) / 2.0

	normPt = coord.Coordinate{
		X: intMidX,
		Y: intMidY,
	}

	n00.X = p1.X - normPt.X
	n00.Y = p1.Y - normPt.Y
	n01.X = p2.X - normPt.X
	n01.Y = p2.Y - normPt.Y
	n10.X = q1.X - normPt.X
	n10.Y = q1.Y - normPt.Y
	n11.X = q2.X - normPt.X
	n11.Y = q2.Y - normPt.Y

	return
}

// safeHCoordinateIntersection computes a segment intersection using homogeneous coordinates.
// Round-off error can cause the raw computation to fail,
// (usually due to the segments being approximately parallel).
// If this happens, a reasonable approximation is computed instead.
func (rli *RobustLineIntersector) safeHCoordinateIntersection(p1, p2, q1, q2 coord.Coordinate) coord.Coordinate {
	intPt, err := hcoordinateIntersection(p1, p2, q1, q2)
	if err != nil {
		intPt = nearestEndpoint(p1, p2, q1, q2)
	}
	return intPt
}

func hcoordinateIntersection(p1, p2, q1, q2 coord.Coordinate) (coord.Coordinate, error) {
	// unrolled computation
	px := p1.Y - p2.Y
	py := p2.X - p1.X
	pw := p1.X*p2.Y - p2.X*p1.Y

	qx := q1.Y - q2.Y
	qy := q2.X - q1.X
	qw := q1.X*q2.Y - q2.X*q1.Y

	x := py*qw - qy*pw
	y := qx*pw - px*qw
	w := px*qy - qx*py

	xInt := x / w
	yInt := y / w

	if math.IsNaN(xInt) || math.IsInf(xInt, 0) || math.IsNaN(yInt) || math.IsInf(yInt, 0) {
		return coord.Coordinate{}, errors.New("Cannot calculate hcoordinate Intersection due to invalid result")
	}

	return coord.Coordinate{
		X: xInt,
		Y: yInt,
	}, nil
}

// nearestEndpoint finds the endpoint of the segments P and Q which
// is closest to the other segment.
// This is a reasonable surrogate for the true
// intersection points in ill-conditioned cases
// (e.g. where two segments are nearly coincident,
// or where the endpoint of one segment lies almost on the other segment).
func nearestEndpoint(p1, p2, q1, q2 coord.Coordinate) coord.Coordinate {
	nearestPt := p1
	minDist := coord.DistancePointToSegment(p1, q1, q2)

	dist := coord.DistancePointToSegment(p2, q1, q2)
	if dist < minDist {
		minDist = dist
		nearestPt = p2
	}
	dist = coord.DistancePointToSegment(q1, p1, p2)
	if dist < minDist {
		minDist = dist
		nearestPt = q1
	}
	dist = coord.DistancePointToSegment(q2, p1, p2)
	if dist < minDist {
		minDist = dist
		nearestPt = q2
	}
	return nearestPt
}

func (rli *RobustLineIntersector) isInSegmentEnvelopes(intPt coord.Coordinate) bool {
	env0 := coord.NewEnvelopeFromCoords(rli.inputLines[0][0], rli.inputLines[0][1])
	env1 := coord.NewEnvelopeFromCoords(rli.inputLines[1][0], rli.inputLines[1][1])
	return env0.ContainsCoord(intPt) && env1.ContainsCoord(intPt)
}
*/
