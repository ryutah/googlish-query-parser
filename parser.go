package gparser

import (
	"bytes"
	"fmt"
	"strings"
)

type NodeType int

const (
	Empty NodeType = iota
	Value
	CompleteMatch
	And
	Or
	Key
)

func Parse(query string) (Node, error) {
	query = strings.TrimSpace(query)
	runes := []rune(query)

	var (
		current          Node = new(EmptyNode)
		token                 = new(bytes.Buffer)
		hasQuota              = false
		parenthesesLevel      = 0
	)
	for i := 0; i < len(runes); i++ {
		switch r := runes[i]; r {
		case ' ':
			if parenthesesLevel > 0 {
				token.WriteRune(r)
				continue
			}

			if hasQuota {
				token.WriteRune(r)
				continue
			}

			if tkn := token.String(); tkn == "or" {
				current = &OrNode{Left: current}
			} else if tkn == "and" {
				current = &AndNode{Left: current}
			} else {
				if !isOperand(current) && current.Type() != Empty {
					current = &AndNode{Left: current}
				}
				node, _ := Parse(token.String())
				current = current.Apply(node)
			}
			token.Reset()
		case '"':
			if parenthesesLevel > 0 {
				token.WriteRune(r)
				continue
			}

			if !hasQuota {
				hasQuota = true
			} else {
				hasQuota = false
				current = current.Apply(&CompleteMatchNode{Value: token.String()})
				token.Reset()
			}
		case ':':
			if parenthesesLevel > 0 {
				token.WriteRune(r)
				continue
			}

			q, n := readUntilFinishLevel(string(runes[i+1:]))
			node, _ := Parse(q)
			fmt.Println(node)
			current = current.Apply(&KeyNode{Key: token.String(), Value: node})
			i = i + n
			token.Reset()
		case '(':
			if parenthesesLevel > 0 {
				token.WriteRune(r)
			}
			parenthesesLevel++
		case ')':
			parenthesesLevel--
			if parenthesesLevel == 0 {
				node, _ := Parse(token.String())
				current = current.Apply(node)
				token.Reset()
			} else {
				token.WriteRune(r)
			}
		default:
			token.WriteRune(r)
		}
	}
	if token.Len() > 0 {
		node := &ValueNode{Value: token.String()}
		current = current.Apply(node)
	}
	return current, nil
}

func readUntilFinishLevel(query string) (result string, n int) {
	grouped := func() (string, int) {
		level := 0
		runes := []rune(query)

		for i := 0; i < len(runes); i++ {
			switch r := runes[i]; r {
			case '(':
				level++
			case ')':
				level--
				if level < 0 {
					return string(runes[1 : i-1]), i
				}
			}
		}
		return "", 0
	}
	others := func() (string, int) {
		runes := []rune(query)
		for i := 0; i < len(runes); i++ {
			if runes[i] == ' ' || runes[i] == ')' {
				return string(runes[:i]), i + 1
			}
		}
		return "", 0
	}

	if strings.HasPrefix(query, "(") {
		return grouped()
	}
	return others()
}

type Visitor interface {
	VisitEmpty(*EmptyNode)
	VisitValue(*ValueNode)
	VisitCompleteMatch(*CompleteMatchNode)
	VisitAnd(*AndNode)
	VisitOr(*OrNode)
	VisitKey(*KeyNode)
}

type Node interface {
	Type() NodeType
	Evaluate(Visitor)
	Apply(Node) Node
	String() string
}

type EmptyNode struct{}

func (n *EmptyNode) Type() NodeType {
	return Empty
}

func (n *EmptyNode) Evaluate(v Visitor) {
	v.VisitEmpty(n)
}

func (n *EmptyNode) Apply(node Node) Node {
	return node
}

func (n *EmptyNode) Create(base Node) Node {
	return n
}

func (n *EmptyNode) String() string {
	return fmt.Sprintf("Empty")
}

type ValueNode struct {
	Value string
}

func (n *ValueNode) Type() NodeType {
	return Value
}

func (n *ValueNode) Evaluate(v Visitor) {
	v.VisitValue(n)
}

func (n *ValueNode) Apply(node Node) Node {
	return &AndNode{
		Left:  n,
		Right: node,
	}
}

func (n *ValueNode) Create(base Node) Node {
	return n
}

func (n *ValueNode) String() string {
	return fmt.Sprintf("Value(%v)", n.Value)
}

type CompleteMatchNode struct {
	Value string
}

func (n *CompleteMatchNode) Type() NodeType {
	return CompleteMatch
}

func (n *CompleteMatchNode) Evaluate(v Visitor) {
	v.VisitCompleteMatch(n)
}

func (n *CompleteMatchNode) Create(base Node) Node {
	return n
}

func (n *CompleteMatchNode) Apply(node Node) Node {
	return &AndNode{
		Left:  n,
		Right: node,
	}
}

func (n *CompleteMatchNode) String() string {
	return fmt.Sprintf("CompleteMatch(%v)", n.Value)
}

type AndNode struct {
	Left  Node
	Right Node
}

func (n *AndNode) Type() NodeType {
	return And
}

func (n *AndNode) Evaluate(v Visitor) {
	v.VisitAnd(n)
}

func (n *AndNode) Apply(node Node) Node {
	if n.Right != nil {
		return &AndNode{
			Left:  n,
			Right: node,
		}
	}
	n.Right = node
	return n
}

func (n *AndNode) Create(base Node) Node {
	return &AndNode{
		Left:  base,
		Right: n.Right,
	}
}

func (n *AndNode) String() string {
	return fmt.Sprintf("And(%v, %v)", n.Left, n.Right)
}

type OrNode struct {
	Left  Node
	Right Node
}

func (n *OrNode) Type() NodeType {
	return Or
}

func (n *OrNode) Evaluate(v Visitor) {
	v.VisitOr(n)
}

func (n *OrNode) Apply(node Node) Node {
	n.Right = node
	return n
}

func (n *OrNode) String() string {
	return fmt.Sprintf("Or(%v, %v)", n.Left, n.Right)
}

type KeyNode struct {
	Key   string
	Value Node
}

func (n *KeyNode) Type() NodeType {
	return Key
}

func (n *KeyNode) Evaluate(v Visitor) {
	v.VisitKey(n)
}

func (n *KeyNode) Apply(node Node) Node {
	n.Value = node
	return n
}

func (n *KeyNode) String() string {
	return fmt.Sprintf("Key(%v: %v)", n.Key, n.Value)
}

func isSurroundedWith(s, surround string) bool {
	return strings.HasPrefix(s, surround) && strings.HasSuffix(s, surround)
}

func isBlank(r rune) bool {
	return r == ' '
}

func isOperand(node Node) bool {
	return node.Type() == And || node.Type() == Or || node.Type() == Key
}
