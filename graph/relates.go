package graph

import (
	"github.com/pkg/errors"
	"github.com/simoncochrane/geoz/coord"
	"github.com/simoncochrane/geoz/geom"
)

const (
	DimensionFalse = -1
)

type IntersectionMatrix [3][3]int

// RelateNode represents a node in the topological graph used to compute spatial relationships.
type RelateNode = Node

func NewRelateNode(point coord.Coordinate, edges *EdgeEndStar) *RelateNode {
	node := NewNode(point, edges)
	rn := RelateNode(*node) // TODO: avoid this copy
	return &rn
}

type Relate struct {
	graphs [2]*Graph
	//geoms  [2]*geom.Geometry

	nodes map[coord.Coordinate]*RelateNode

	lineIntersector coord.LineIntersector
	pointLocator    *PointLocator
}

func NewIntersectionMatrix() IntersectionMatrix {
	var im IntersectionMatrix
	im.SetAll(DimensionFalse)
	return im
}

func (im IntersectionMatrix) SetAll(dimVal int) {
	for i := 0; i < len(im); i++ {
		for j := 0; j < len(im[i]); j++ {
			im[i][j] = dimVal
		}
	}
}

func (im IntersectionMatrix) Set(row, col, dim int) {
	im[row][col] = dim
}

func NewRelate(a, b *geom.Geometry) (*Relate, error) {
	ga, err := NewGraph(a, 0, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create geometry graph for 1st Geometry")
	}

	gb, err := NewGraph(b, 1, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create geometry graph for 2nd Geometry")
	}

	return &Relate{
		graphs: [2]*Graph{
			ga, gb,
		},
		lineIntersector: coord.NewRobustLineIntersector(),
	}, nil
}

func (r *Relate) IntersectionMatrix() IntersectionMatrix {
	im := NewIntersectionMatrix()
	im.Set(coord.LocationExterior, coord.LocationExterior, 2)

	if !r.graphs[0].geometry.Envelope().Intersects(r.graphs[1].geometry.Envelope()) {
		r.computeDisjointIM(im)
		return im
	}

	r.graphs[0].computeSelfNodes(r.lineIntersector, !r.graphs[0].geometry.IsRings(), false)
	r.graphs[1].computeSelfNodes(r.lineIntersector, !r.graphs[1].geometry.IsRings(), false)

	// compute intersections between edges of the two input geometries
	intersector := r.graphs[0].computeEdgeIntersections(r.graphs[1], r.lineIntersector, false)

	r.computeIntersectionNodes(0)
	r.computeIntersectionNodes(1)

	// Copy the labelling for the nodes in the parent Geometries.  These override
	// any labels determined by intersections between the geometries.
	r.copyNodesAndLabels(0)
	r.copyNodesAndLabels(1)

	// complete the labelling for any nodes which only have a label for a single geometry
	r.labelIsolatedNodes()

	/*

	    // If a proper intersection was found, we can set a lower bound on the IM.
	    computeProperIntersectionIM(intersector, im);

	    /**
	     * Now process improper intersections
	     * (eg where one or other of the geometries has a vertex at the intersection point)
	     * We need to compute the edge graph at all nodes to determine the IM.
	     *

	    // build EdgeEnds for all intersections
	    EdgeEndBuilder eeBuilder = new EdgeEndBuilder();
	    List ee0 = eeBuilder.computeEdgeEnds(arg[0].getEdgeIterator());
	    insertEdgeEnds(ee0);
	    List ee1 = eeBuilder.computeEdgeEnds(arg[1].getEdgeIterator());
	    insertEdgeEnds(ee1);

	    labelNodeEdges();

	  /**
	   * Compute the labeling for isolated components
	   * <br>
	   * Isolated components are components that do not touch any other components in the graph.
	   * They can be identified by the fact that they will
	   * contain labels containing ONLY a single element, the one for their parent geometry.
	   * We only need to check components contained in the input graphs, since
	   * isolated components will not have been replaced by new components formed by intersections.
	   *
	    labelIsolatedEdges(0, 1);
	    labelIsolatedEdges(1, 0);

	    // update the IM from all components
	    updateIM(im);
	    return im;
	*/
}

func (r *Relate) computeDisjointIM(im IntersectionMatrix) {
	if !r.graphs[0].geometry.IsEmpty() {
		im.Set(coord.LocationInterior, coord.LocationExterior, r.graphs[0].geometry.Dimension())
		im.Set(coord.LocationBoundary, coord.LocationExterior, r.graphs[0].geometry.BoundaryDimension())
	}
	if !r.graphs[1].geometry.IsEmpty() {
		im.Set(coord.LocationExterior, coord.LocationInterior, r.graphs[1].geometry.Dimension())
		im.Set(coord.LocationExterior, coord.LocationBoundary, r.graphs[1].geometry.BoundaryDimension())
	}
}

func (r *Relate) computeIntersectionNodes(argIndex int) {
	for _, edge := range r.graphs[argIndex].edges {
		edgeLoc := edge.Label.LocationAt(argIndex, PositionOn)

		for _, edgeInt := range edge.eiList.nodeMap {
			relateNode, has := r.nodes[edgeInt.Coordinate]
			if !has {
				relateNode = NewRelateNode(edgeInt.Coordinate, NewEdgeEndStar())
				r.nodes[edgeInt.Coordinate] = relateNode
			}
			if relateNode.Label.IsNil(argIndex) {
				if edgeLoc == coord.LocationBoundary {
					relateNode.SetLabelBoundary(argIndex)
				} else {
					relateNode.Label.SetLocation(argIndex, coord.LocationInterior)
				}
			}
		}
	}
}

func (r *Relate) copyNodesAndLabels(argIndex int) {
	for _, node := range r.graphs[argIndex].nodes {
		newNode := NewNode(node.Point, NewEdgeEndStar())
		newNode.Label.SetLocation(argIndex, node.Label.LocationAt(argIndex, PositionOn))
		r.nodes[node.Point] = newNode
	}
}

func (r *Relate) labelIsolatedNodes() {
	for _, node := range r.nodes {
		if node.Label.GeometryCount() == 0 {
			panic("Expected label geometry count to be > 0")
		}
		if node.Isolated() {
			if node.Label.IsNil(0) {
				r.labelIsolatedNode(node, 0)
			} else {
				r.labelIsolatedNode(node, 1)
			}
		}
	}
}

func (r *Relate) labelIsolatedNode(node *Node, targetIndex int) {
	loc := r.pointLocator.Locate(node.Point, r.graphs[targetIndex].geometry)
	node.Label.SetAllLocations(targetIndex, loc)

	/////

	/*
	   int loc = ptLocator.locate(n.getCoordinate(), arg[targetIndex].getGeometry());
	   n.getLabel().setAllLocations(targetIndex, loc);
	*/
}
