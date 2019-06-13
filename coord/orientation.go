package coord

import (
	"errors"
	"math/big"
)

const (
	DPSafeEpsilon = 1e-15
)

// IsCCW returns true if the given ring is oriented counter clockwise.
// Returns an error if the coordinates do not represent a valid ring.
func IsCCW(ring Coordinates) (bool, error) {
	// number of points without closing endpoint
	nPts := len(ring) - 1

	if nPts < 3 {
		return false, errors.New("ring must have at least 3 points to determine orientation")
	}

	highest := ring[0]
	highestIdx := 0
	for i, coord := range ring {
		if coord.Y > highest.Y {
			highest = coord
			highestIdx = i
		}
	}

	// find distinct point before highest
	iPrev := highestIdx - 1
	if iPrev < 0 {
		iPrev = nPts
	}
	for ring[iPrev].Equals2D(highest) && iPrev != highestIdx {
		iPrev--
		if iPrev < 0 {
			iPrev = nPts
		}
	}

	// find distinct point after highest
	iNext := (highestIdx + 1) % nPts
	for ring[iNext].Equals2D(highest) && iNext != highestIdx {
		iNext = (iNext + 1) % nPts
	}

	prev := ring[iPrev]
	next := ring[iNext]

	// catch edge cases (where there aren't 3 distinct points)
	if prev.Equals2D(highest) || next.Equals2D(highest) || prev.Equals2D(next) {
		return false, nil
	}

	disc := OrientationIndex(prev, highest, next)

	if disc == 0 {
		return prev.X > next.X, nil
	}
	return disc > 0, nil
}

// OrientationIndex returns the orientation index of the direction of the point q
// relative to a directed infinite line specified by p1-p2.
func OrientationIndex(p1, p2, q Coordinate) int {
	// attempt fast filter to avoid high precision arithmetic
	index := orientationIndexFilter(p1, p2, q)
	if index <= 1 {
		return index
	}

	var dx1, dy1, dx2, dy2 big.Float
	var p1x, p1y, p2x, p2y big.Float

	dx1.SetFloat64(p2.X)
	p1x.SetFloat64(-p1.X)
	dx1.Add(&dx1, &p1x)

	dy1.SetFloat64(p2.Y)
	p1y.SetFloat64(-p1.Y)
	dy1.Add(&dy1, &p1y)

	dx2.SetFloat64(q.X)
	p2x.SetFloat64(-p2.X)
	dx2.Add(&dx2, &p2x)

	dy2.SetFloat64(q.Y)
	p2y.SetFloat64(-p2.Y)
	dy2.Add(&dy2, &p2y)

	dx1.Mul(&dx1, &dy2)
	dy1.Mul(&dy1, &dx2)
	dx1.Sub(&dx1, &dy1)

	return dx1.Sign()
}

func orientationIndexFilter(pa, pb, pc Coordinate) int {
	var detsum float64

	detleft := (pa.X - pc.X) * (pb.Y - pc.Y)
	detright := (pa.Y - pc.Y) * (pb.X - pc.X)
	det := detleft - detright

	if detleft > 0.0 {
		if detright <= 0.0 {
			return signum(det)
		} else {
			detsum = detleft + detright
		}
	} else if detleft < 0.0 {
		if detright >= 0.0 {
			return signum(det)
		} else {
			detsum = -detleft - detright
		}
	} else {
		return signum(det)
	}

	errbound := DPSafeEpsilon * detsum
	if det >= errbound || -det >= errbound {
		return signum(det)
	}

	return 2
}

func signum(x float64) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}
