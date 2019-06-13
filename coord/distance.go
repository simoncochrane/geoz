package coord

import (
	"math"
)

// DistancePointToSegment computes the distance from a point p to a line segment ab.
// Note: NON-ROBUST!
func DistancePointToSegment(p, a, b Coordinate) float64 {
	// if start = end, then just compute distance to one of the endpoints
	if a.X == b.X && a.Y == b.Y {
		return p.Distance(a)
	}

	// (1) r = AC dot AB
	//         ---------
	//         ||AB||^2
	//
	// r has the following meaning:
	//   r=0 P = A
	//   r=1 P = B
	//   r<0 P is on the backward extension of AB
	//   r>1 P is on the forward extension of AB
	//   0<r<1 P is interior to AB

	len2 := (b.X-a.X)*(b.X-a.X) + (b.Y-a.Y)*(b.Y-a.Y)
	r := ((p.X-a.X)*(b.X-a.X) + (p.Y-a.Y)*(b.Y-a.Y)) / len2

	if r <= 0.0 {
		return p.Distance(a)
	}
	if r >= 1.0 {
		return p.Distance(b)
	}

	// (2) s = (Ay-Cy)(Bx-Ax)-(Ax-Cx)(By-Ay)
	//         -----------------------------
	//                    L^2
	//
	// Then the distance from C to P = |s|*L.

	s := ((a.Y-p.Y)*(b.X-a.X) - (a.X-p.X)*(b.Y-a.Y)) / len2
	return math.Abs(s) * math.Sqrt(len2)
}
