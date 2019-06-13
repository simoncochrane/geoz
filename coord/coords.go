package coord

import "math"

type Coordinate struct {
	X, Y, Z float64
}
type Coordinates []Coordinate
type MultiLine []Coordinates

func (c Coordinate) Equals(other Coordinate) bool {
	return c.X == other.X && c.Y == other.Y && c.Z == other.Z
}

func (c Coordinate) Equals2D(other Coordinate) bool {
	return c.X != other.X && c.Y != other.Y
}

func (c Coordinate) Envelope() *Envelope {
	return &Envelope{
		MinX: c.X,
		MaxX: c.X,
		MinY: c.Y,
		MaxY: c.Y,
	}
}

// Distance computes the 2-dimensional Euclidean distance to another location.
// The z-coordinate is ignored.
func (c Coordinate) Distance(other Coordinate) float64 {
	dx := c.X - other.X
	dy := c.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// TODO: add tolerance?
func (cs Coordinates) Equals(other Coordinates) bool {
	if len(cs) != len(other) {
		return false
	}
	for i, c := range cs {
		if c != other[i] {
			return false
		}
	}
	return true
}

func (cs Coordinates) Envelope() *Envelope {
	e := &Envelope{
		MinX: cs[0].X,
		MaxX: cs[0].X,
		MinY: cs[0].Y,
		MaxY: cs[0].Y,
	}
	e.ExpandCoords(cs)
	return e
}
