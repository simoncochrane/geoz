package coord

type Location int

const (
	LocationNone     Location = -1
	LocationInterior          = 0
	LocationBoundary          = 1
	LocationExterior          = 2
)

// PointOnLine tests whether the point lies on the given line.
func PointOnLine(point Coordinate, line []Coordinate) bool {
	lineIntersector := NewRobustLineIntersector()

	for i := 1; i < len(line); i++ {
		lineIntersector.ComputePointIntersection(point, line[i-1], line[i])
		if lineIntersector.HasIntersection() {
			return true
		}
	}

	return false
}

func PointInRing(point Coordinate, ring []Coordinate) Location {
	// TODO:
	//    return RayCrossingCounter.locatePointInRing(p, ring);

}
