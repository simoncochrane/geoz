package graph

import (
	"github.com/simoncochrane/geoz/coord"
	"github.com/simoncochrane/geoz/geom"
)

type PointLocator struct {
	boundaryNodeRule BoundaryNodeRule

	isIn          bool
	numBoundaries int
}

func NewPointLocator() *PointLocator {
	return &PointLocator{
		boundaryNodeRule: NewMod2BoundaryNodeRule(),
	}
}

func (pl *PointLocator) Locate(point coord.Coordinate, geometry *geom.Geometry) coord.Location {
	if geometry.IsEmpty() {
		return coord.LocationExterior
	}

	if geometry.Type == geom.TypeLineString {
		return pl.locateOnLineString(point, geometry)
	} else if geometry.Type == geom.TypePolygon {
		return pl.locateInPolygon(point, geometry)
	}

	pl.isIn = false
	pl.numBoundaries = 0
	pl.computeLocation(point, geometry)

	if pl.boundaryNodeRule.InBoundary(pl.numBoundaries) {
		return coord.LocationBoundary
	}
	if pl.numBoundaries > 0 || pl.isIn {
		return coord.LocationInterior
	}
	return coord.LocationExterior
}

func (pl *PointLocator) computeLocation(point coord.Coordinate, geometry *geom.Geometry) {
	switch geometry.Type {
	case geom.TypePoint:
		pl.updateLocationInfo(pl.locateOnPoint(point, geometry))
	case geom.TypeLineString:
		pl.updateLocationInfo(pl.locateOnLineString(point, geometry))
	case geom.TypePolygon:
		pl.updateLocationInfo(pl.locateInPolygon(point, geometry))
		// TODO:
	case geom.TypeMultiLineString:
		// TODO:
	case geom.TypeMultiPolygon:
		// TODO:
	case geom.TypeCollection:
		// TODO:
	}
}

func (pl *PointLocator) updateLocationInfo(location coord.Location) {
	if location == coord.LocationInterior {
		pl.isIn = true
	}
	if location == coord.LocationBoundary {
		pl.numBoundaries++
	}
}

func (pl *PointLocator) locateOnPoint(point coord.Coordinate, geometry *geom.Geometry) coord.Location {
	if geometry.Coord.Equals2D(point) {
		return coord.LocationInterior
	}
	return coord.LocationExterior
}

func (pl *PointLocator) locateOnLineString(point coord.Coordinate, geometry *geom.Geometry) coord.Location {
	// bounding box check
	if !geometry.Line.Envelope().IntersectsPoint(point) {
		return coord.LocationExterior
	}

	if !geometry.IsClosed() {
		if point.Equals(geometry.Line[0]) || point.Equals(geometry.Line[len(geometry.Line)-1]) {
			return coord.LocationBoundary
		}
	}

	if coord.PointOnLine(point, geometry.Line) {
		return coord.LocationInterior
	}
	return coord.LocationExterior
}

func (pl *PointLocator) locateInPolygon(point coord.Coordinate, geometry *geom.Geometry) coord.Location {
	if geometry.IsEmpty() {
		return coord.LocationExterior
	}

	shell := geometry.Line

	shellLoc := pl.locateInPolygonRing(point, shell)

	/////

	/*
	   if (poly.isEmpty()) return Location.EXTERIOR;

	   LinearRing shell = (LinearRing) poly.getExteriorRing();

	   int shellLoc = locateInPolygonRing(p, shell);
	   if (shellLoc == Location.EXTERIOR) return Location.EXTERIOR;
	   if (shellLoc == Location.BOUNDARY) return Location.BOUNDARY;
	   // now test if the point lies in or on the holes
	   for (int i = 0; i < poly.getNumInteriorRing(); i++) {
	     LinearRing hole = (LinearRing) poly.getInteriorRingN(i);
	     int holeLoc = locateInPolygonRing(p, hole);
	     if (holeLoc == Location.INTERIOR) return Location.EXTERIOR;
	     if (holeLoc == Location.BOUNDARY) return Location.BOUNDARY;
	   }
	   return Location.INTERIOR;
	*/
}

func (pl *PointLocator) locateInPolygonRing(point coord.Coordinate, ring coord.Coordinates) coord.Location {
	// bounding box check
	if !ring.Envelope().IntersectsPoint(point) {
		return coord.LocationExterior
	}

	return coord.PointInRing(point, ring)
}
