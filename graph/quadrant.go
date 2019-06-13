package graph

import (
	"github.com/pkg/errors"
	"github.com/simoncochrane/geoz/coord"
)

type Quadrant int

const (
	QuadrantNE Quadrant = 0
	QuadrantNW          = 1
	QuadrantSW          = 2
	QuadrantSE          = 3
)

func QuadrantCoord(p0, p1 coord.Coordinate) (Quadrant, error) {
	if p0.X == p1.X && p0.Y == p1.Y {
		return 0, errors.Errorf("Cannot compute quadrant for two identical points: %v", p0)
	}

	if p1.X >= p0.X {
		if p1.Y >= p0.Y {
			return QuadrantNE, nil
		} else {
			return QuadrantSE, nil
		}
	} else {
		if p1.Y >= p0.Y {
			return QuadrantNW, nil
		} else {
			return QuadrantSW, nil
		}
	}
}
