package coord

import (
	"math"
)

type Envelope struct {
	MinX, MaxX, MinY, MaxY float64
}

func NewEnvelope(minX, maxX, minY, maxY float64) *Envelope {
	e := &Envelope{}
	if minX < maxX {
		e.MinX = minX
		e.MaxX = maxX
	} else {
		e.MinX = maxX
		e.MaxX = minX
	}
	if minY < maxY {
		e.MinY = minY
		e.MaxY = maxY
	} else {
		e.MinY = maxY
		e.MaxY = minY
	}
	return e
}

func NewEnvelopeFromCoords(p1, p2 Coordinate) *Envelope {
	return NewEnvelope(p1.X, p1.Y, p2.X, p2.Y)
}

func (e *Envelope) Expand(c Coordinate) {
	if c.X < e.MinX {
		e.MinX = c.X
	}
	if c.X > e.MaxX {
		e.MaxX = c.X
	}
	if c.Y < e.MinY {
		e.MinY = c.Y
	}
	if c.Y > e.MaxY {
		e.MaxY = c.Y
	}
}

func (e *Envelope) ExpandCoords(cs Coordinates) {
	for _, c := range cs {
		e.Expand(c)
	}
}

func (e *Envelope) ExpandEnvelope(other *Envelope) {
	if other.MinX < e.MinX {
		e.MinX = other.MinX
	}
	if other.MaxX > e.MaxX {
		e.MaxX = other.MaxX
	}
	if other.MinY < e.MinY {
		e.MinY = other.MinY
	}
	if other.MaxY > e.MaxY {
		e.MaxY = other.MaxY
	}
}

func (e *Envelope) Intersects(other *Envelope) bool {
	if e == nil || other == nil {
		return false
	}
	return !(other.MinX > e.MaxX ||
		other.MaxX < e.MinX ||
		other.MinY > e.MaxY ||
		other.MaxY < e.MinY)
}

func (e *Envelope) IntersectsPoint(point Coordinate) bool {
	return e.IntersectsXY(point.X, point.Y)
}

func (e *Envelope) IntersectsXY(x, y float64) bool {
	if e == nil {
		return false
	}
	return !(x > e.MaxX || x < e.MinX || y > e.MaxY || y < e.MinY)
}

func (e *Envelope) Contains(x, y float64) bool {
	if e == nil {
		return false
	}
	return x >= e.MinX && x <= e.MaxX && y >= e.MinY && y <= e.MaxY
}

func (e *Envelope) ContainsCoord(p Coordinate) bool {
	return e.Contains(p.X, p.Y)
}

// EnvelopeIntersects returns true if the envelope defined by p1-p2 and the
// envelope defined by q1-q2 intersect.
func EnvelopeIntersects(p1, p2, q1, q2 Coordinate) bool {
	minp := math.Min(p1.X, p2.X)
	maxp := math.Max(p1.X, p2.X)
	minq := math.Min(q1.X, q2.X)
	maxq := math.Max(q1.X, q2.X)

	if minp > maxq {
		return false
	}
	if maxp < minq {
		return false
	}

	minp = math.Min(p1.Y, p2.Y)
	maxp = math.Max(p1.Y, p2.Y)
	minq = math.Min(q1.Y, q2.Y)
	maxq = math.Max(q1.Y, q2.Y)

	if minp > maxq {
		return false
	}
	if maxp < minq {
		return false
	}
	return true
}

// EnvelopeIntersectsPoint tests whether the point q intersects with the Envelope defined by p1-p2.
func EnvelopeIntersectsPoint(p1, p2, q Coordinate) bool {
	minpx := math.Min(p1.X, p2.X)
	maxpx := math.Max(p1.X, p2.X)
	minpy := math.Min(p1.Y, p2.Y)
	maxpy := math.Max(p1.Y, p2.Y)

	if q.X >= minpx && q.X <= maxpx && q.Y >= minpy && q.Y <= maxpy {
		return true
	}
	return false
}
