package fp

import (
	"fmt"
	"strings"
)

type NodeType int

const (
	NodeGroup NodeType = iota
	NodeComparison
	NodeLogical
)

type Node struct {
	Type       NodeType
	Nodes      []*Node // If Group
	Comparison *Comparison
	Logic      Logic
}

func (n Node) String() string {
	switch n.Type {
	case NodeGroup:
		s := ""
		for _, child := range n.Nodes {
			s += child.String()
		}
		return s
	case NodeComparison:
		return n.Comparison.String()
	case NodeLogical:
		return n.Logic.String()
	}

	return "unknown_node_type"
}

func newNode(typ NodeType) *Node {
	return &Node{
		Type:  typ,
		Nodes: []*Node{},
	}
}

func newCompNode(comp *Comparison) *Node {
	return &Node{
		Type:       NodeComparison,
		Nodes:      []*Node{},
		Comparison: comp,
	}
}

func newLogicNode(l Logic) *Node {
	return &Node{
		Type:  NodeLogical,
		Nodes: []*Node{},
		Logic: l,
	}
}

type Comparison struct {
	Selector  string
	Operator  Expr
	Arguments []string
}

func (c Comparison) String() string {
	args := ""
	if len(c.Arguments) > 1 {
		args = "('" + strings.Join(c.Arguments, "','") + "')"
	} else {
		args = "'" + c.Arguments[0] + "'"
	}
	return fmt.Sprintf("%s=%s=%s", c.Selector, c.Operator, args)
}

func newComparison(sel string) *Comparison {
	return &Comparison{
		Selector:  sel,
		Arguments: []string{},
	}
}

type Logic int

const (
	And Logic = iota
	Or
)

func (l Logic) String() string {
	switch l {
	case And:
		return ";"
	case Or:
		return ","
	default:
		panic(fmt.Errorf("invalid logic operator: %d", l))
	}
}
