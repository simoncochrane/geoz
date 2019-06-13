package graph

type BoundaryNodeRule interface {
	InBoundary(boundaryCount int) bool
}

type Mod2BoundaryNodeRule struct{}

func NewMod2BoundaryNodeRule() *Mod2BoundaryNodeRule {
	return &Mod2BoundaryNodeRule{}
}

func (bnr *Mod2BoundaryNodeRule) InBoundary(boundaryCount int) bool {
	return boundaryCount%2 == 1
}
