package graph

import (
	"sort"

	"github.com/pkg/errors"
)

type EventType int

const (
	EventTypeInsert EventType = 1
	EventTypeDelete           = 2
)

type EdgeSetIntersector interface {
	// ComputeSelfIntersections computes all self-intersections between edges in a set of edges,
	// allowing client to choose whether self-intersections are computed.
	ComputeSelfIntersections(edges []*Edge, si *SegmentIntersector, computeAllSegments bool) error

	// ComputeEdgeIntersections computes all mutual intersections between two sets of edges.
	ComputeEdgeIntersections(edges0, edges1 []*Edge, si *SegmentIntersector) error
}

type SweepLineEvent struct {
	Type          EventType
	Label         interface{} // used for red-blue intersection detection
	X             float64
	MonotoneChain *MonotoneChain

	// null if this is an insert event
	InsertEvent      *SweepLineEvent
	DeleteEventIndex int
}

type SweepLineEvents []*SweepLineEvent

func (sle *SweepLineEvent) HasSameLabel(other *SweepLineEvent) bool {
	if sle.Label == nil {
		return false
	}
	return sle.Label == other.Label
}

func (sles SweepLineEvents) Len() int      { return len(sles) }
func (sles SweepLineEvents) Swap(i, j int) { sles[i], sles[j] = sles[j], sles[i] }
func (sles SweepLineEvents) Less(i, j int) bool {
	if sles[i].X < sles[j].X {
		return true
	}
	if sles[i].X == sles[j].X {
		return sles[i].Type < sles[j].Type
	}
	return false
}

type SimpleMCSweepLineIntersector struct {
	events      SweepLineEvents
	numOverlaps int
}

func NewSimpleMCSweepLineIntersector() *SimpleMCSweepLineIntersector {
	return &SimpleMCSweepLineIntersector{}
}

func (sli *SimpleMCSweepLineIntersector) ComputeSelfIntersections(edges []*Edge,
	si *SegmentIntersector, computeAllSegments bool) error {

	var err error
	if computeAllSegments {
		err = sli.addEdgesWithEdgeSet(edges, nil)
	} else {
		err = sli.addEdges(edges)
	}

	if err != nil {
		return errors.Wrap(err, "failed to add edges")
	}

	sli.computeIntersections(si)
	return nil
}

func (sli *SimpleMCSweepLineIntersector) ComputeEdgeIntersections(edges0, edges1 []*Edge, si *SegmentIntersector) error {
	if err := sli.addEdgesWithEdgeSet(edges0, edges0); err != nil {
		return errors.Wrap(err, "failed to add edges0")
	}
	if err := sli.addEdgesWithEdgeSet(edges1, edges1); err != nil {
		return errors.Wrap(err, "failed to add edges1")
	}

	sli.computeIntersections(si)
	return nil
}

func (sli *SimpleMCSweepLineIntersector) addEdges(edges []*Edge) error {
	for _, edge := range edges {
		if err := sli.addEdge(edge, edge); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (sli *SimpleMCSweepLineIntersector) addEdgesWithEdgeSet(edges []*Edge, edgeSet interface{}) error {
	for _, edge := range edges {
		if err := sli.addEdge(edge, edgeSet); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (sli *SimpleMCSweepLineIntersector) addEdge(edge *Edge, edgeSet interface{}) error {
	mce, err := edge.MonotoneChainEdge()
	if err != nil {
		return errors.WithStack(err)
	}

	for i := range mce.StartIndexes {
		mc := mce.MonotoneChain(i)
		insertEvent := &SweepLineEvent{
			Type:          EventTypeInsert,
			Label:         edgeSet,
			X:             mce.MinX(i),
			MonotoneChain: mc,
		}
		deleteEvent := &SweepLineEvent{
			Type:        EventTypeDelete,
			X:           mce.MaxX(i),
			InsertEvent: insertEvent,
		}

		sli.events = append(sli.events, insertEvent, deleteEvent)
	}

	return nil
}

func (sli *SimpleMCSweepLineIntersector) computeIntersections(si *SegmentIntersector) {
	sli.numOverlaps = 0
	sli.prepareEvents()

	for i, ev := range sli.events {
		if ev.Type == EventTypeInsert {
			sli.processOverlaps(i, ev.DeleteEventIndex, ev, si)
		}
		if si.Done() {
			break
		}
	}
}

func (sli *SimpleMCSweepLineIntersector) prepareEvents() {
	sort.Sort(sli.events)

	for i, ev := range sli.events {
		if ev.Type == EventTypeDelete {
			ev.InsertEvent.DeleteEventIndex = i
		}
	}
}

func (sli *SimpleMCSweepLineIntersector) processOverlaps(start, end int, event *SweepLineEvent, si *SegmentIntersector) {
	// since we might need to test for self-intersections, include current INSERT event in list
	// of event objects to test.
	// Last index can be skipped because it must be a Delete event.
	for _, ev := range sli.events {
		if ev.Type == EventTypeInsert {
			if !event.HasSameLabel(ev) {
				event.MonotoneChain.ComputeIntersections(ev.MonotoneChain, si)
				sli.numOverlaps++
			}
		}
	}
}
