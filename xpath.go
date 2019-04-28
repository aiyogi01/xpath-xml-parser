// The purpose of this file is to define neccessary abstactions for
// using XPath queries which hide the underlying implementation specific
// to the used XPath library. In this way, the underlying XPath library
// can be easliy exchanged without introducing breaking changes in the code
// that depends on these abstactions.
//
// The two main types with their interfaces defined here are:
//
//   * Node  - which represents a node in an XML document
//   * XPath - which represents a compiled XPath expression

package main

import (
	"io"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
)

// ---------------------------------------------------------------------------
//	Node
// ---------------------------------------------------------------------------

// Node represents a node in an XML document
type Node struct {
	data *xmlquery.Node
}

// ParseXML parses an XML file and returns the root node
func ParseXML(r io.Reader) (*Node, error) {
	root, err := xmlquery.Parse(r)
	if err != nil {
		return nil, err
	}
	return &Node{data: root}, nil
}

// InnerText selects the inner text of the node
func (node *Node) InnerText() string {
	return node.data.InnerText()
}

// InnerXML selectes the raw inner XML string of the node
func (node *Node) InnerXML() string {
	return node.data.OutputXML(false)
}

// ---------------------------------------------------------------------------
//	XPath
// ---------------------------------------------------------------------------

// XPath is a compiled XPath expression
type XPath struct {
	data *xpath.Expr
}

// CompileXPath takes a string and constructs a compiled XPath expression
func CompileXPath(expr string) (*XPath, error) {
	compiled, err := xpath.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &XPath{data: compiled}, err
}

// Find applies the XPath expression to the given node
// and returns a list of all matched descendant nodes
func (xpath *XPath) Find(node *Node) []*Node {
	return xpath.find(node)
}

// FindOne applies the XPath expression to the given node
// and returns the first matched descendant node
func (xpath *XPath) FindOne(node *Node) *Node {
	return xpath.findOne(node)
}

// find implements the method Find and
// is specific to the underlying XPath library
func (xpath *XPath) find(node *Node) []*Node {
	iter := xpath.data.Select(xmlquery.CreateXPathNavigator(node.data))
	var matches []*Node
	for iter.MoveNext() {
		data := getCurrentNode(iter)
		matches = append(matches, &Node{data: data})
	}
	return matches
}

// findOne implements the method FindOne and
// is specific to the underlying XPath library
func (xpath *XPath) findOne(node *Node) *Node {
	iter := xpath.data.Select(xmlquery.CreateXPathNavigator(node.data))
	var match *Node
	if iter.MoveNext() {
		data := getCurrentNode(iter)
		match = &Node{data: data}
	}
	return match
}

// getCurrentNode as a not exported function has been taken unaltered
// from the the file "query.go" in the package "github.com/antchfx/xmlquery"
func getCurrentNode(it *xpath.NodeIterator) *xmlquery.Node {
	n := it.Current().(*xmlquery.NodeNavigator)
	if n.NodeType() == xpath.AttributeNode {
		childNode := &xmlquery.Node{
			Type: xmlquery.TextNode,
			Data: n.Value(),
		}
		return &xmlquery.Node{
			Type:       xmlquery.AttributeNode,
			Data:       n.LocalName(),
			FirstChild: childNode,
			LastChild:  childNode,
		}
	}
	return n.Current()
}
