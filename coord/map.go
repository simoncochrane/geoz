package coord

import "math"

// coordsEntry contains the coordinates key and the value added to the map.
type coordsEntry struct {
	Coordinates
	Value interface{}
}

type CoordinatesMap struct {
	entries map[uint64][]*coordsEntry
}

func NewCoordinatesMap() *CoordinatesMap {
	return &CoordinatesMap{
		entries: map[uint64][]*coordsEntry{},
	}
}

func (cm *CoordinatesMap) Add(coords Coordinates, value interface{}) {
	hc := hashCodeCoordinates(coords)
	vals, has := cm.entries[hc]
	if !has {
		cm.entries[hc] = []*coordsEntry{&coordsEntry{coords, value}}
		return
	}

	for _, val := range vals {
		if coords.Equal(val.Coordinates) {
			val.Value = value
			return
		}
	}

	cm.entries[hc] = append(vals, &coordsEntry{coords, value})
}

func (cm *CoordinatesMap) Get(coords Coordinates) (interface{}, bool) {
	hc := hashCodeCoordinates(coords)
	vals, has := cm.entries[hc]
	if !has {
		return nil, false
	}

	for _, val := range vals {
		if coords.Equal(val.Coordinates) {
			return val.Value, true
		}
	}

	return nil, false
}

func hashCodeFloat64(x float64) uint64 {
	f := math.Float64bits(x)
	return f ^ (f >> 32)
}

func hashCodeCoordinate(c Coordinate) uint64 {
	result := uint64(17)
	result = 37*result + hashCodeFloat64(c.X)
	result = 37*result + hashCodeFloat64(c.Y)
	return result
}

func hashCodeCoordinates(coords Coordinates) uint64 {
	e := coords.Envelope()
	return hashCodeEnvelope(e)
}

func hashCodeEnvelope(e *Envelope) uint64 {
	result := uint64(17)
	result = 37*result + hashCodeFloat64(e.MinX)
	result = 37*result + hashCodeFloat64(e.MaxX)
	result = 37*result + hashCodeFloat64(e.MinY)
	result = 37*result + hashCodeFloat64(e.MaxY)
	return result
}
