package operation

import "github.com/simoncochrane/geoz/geom"

type IntersectionMatrix [][]int

type Relate struct {
}

func NewRelate(a, b *geom.Geometry) *Relate {
	return &Relate{}
}

func (r *Relate) IntersectionMatrix() IntersectionMatrix {

}
