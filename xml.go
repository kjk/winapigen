package main

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// Using tree builder from
// https://stackoverflow.com/questions/30256729/how-to-traverse-through-xml-data-in-golang

// XMLNode describes a node in XML tree
type XMLNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []*XMLNode `xml:",any"`
}

// UnmarshalXML recursively decodes XML into a tree
func (n *XMLNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node XMLNode
	return d.DecodeElement((*node)(n), &start)
}

func (n *XMLNode) childByName(childLocalName string) *XMLNode {
	for _, n = range n.Nodes {
		if n.XMLName.Local == childLocalName {
			return n
		}
	}
	return nil
}

func (n *XMLNode) String() string {
	s := serAttributes(n.Attrs)
	start := "<"
	end := "/>"
	if len(n.Nodes) > 0 {
		end = ">"
	}
	if len(s) > 0 {
		return start + n.XMLName.Local + " " + s + end
	}
	return start + n.XMLName.Local + s + end
}

func hasChildNode(n *XMLNode, childLocalName string) bool {
	if n.childByName(childLocalName) == nil {
		return false
	}
	return true
}

func serAttributes(attrs []xml.Attr) string {
	if len(attrs) == 0 {
		return ""
	}
	var a []string
	for _, attr := range attrs {
		s := fmt.Sprintf("%s=%s", attr.Name.Local, esc(attr.Value))
		a = append(a, s)
	}
	return strings.Join(a, " ")
}

func esc(s string) string {
	if strings.Contains(s, " ") {
		return fmt.Sprintf(`"%s"`, s)
	}
	return s
}

func xmlSerializeNodeRecur(node *XMLNode, level int) []string {
	indent := strings.Repeat("  ", level)
	if len(node.Nodes) == 0 {
		return []string{indent + node.String()}
	}
	lines := []string{indent + node.String()}
	for _, n := range node.Nodes {
		childLines := xmlSerializeNodeRecur(n, level+1)
		lines = append(lines, childLines...)
	}
	s := fmt.Sprintf("%s</%s>", indent, node.XMLName.Local)
	return append(lines, s)
}

func xmlSerializeNode(node *XMLNode, level int) string {
	lines := xmlSerializeNodeRecur(node, level)
	return strings.Join(lines, "\n")
}

func xmlPrintFunc(node *XMLNode, level int) bool {
	fmt.Printf("%s%s\n", strings.Repeat("  ", level), node)
	return true
}

func xmlPrint(root *XMLNode) {
	nodes := []*XMLNode{root}
	xmlWalk(nodes, 0, xmlPrintFunc)
}

func xmlWalk(nodes []*XMLNode, level int, f func(*XMLNode, int) bool) {
	for _, n := range nodes {
		if f(n, level) {
			xmlWalk(n.Nodes, level+1, f)
		}
	}
}

func mustNoAttrs(attrs []xml.Attr, n *XMLNode) {
	if n != nil {
		//panicIf(len(attrs) != 0, "unsupported attribute in node '%s'", xmlSerializeNode(n, 0))
		panicIf(len(attrs) != 0, "unsupported attribute in node '%s'", n)
	} else {
		panicIf(len(attrs) != 0, "unsupported attribute: '%s'", serAttributes(attrs))
	}
}

func mustNoChildren(n *XMLNode) {
	panicIf(len(n.Nodes) > 0, "unexpected children in node:\n%s\n", xmlSerializeNode(n, 0))
}

func mustTag(is string, wanted string) {
	if is != wanted {
		err := fmt.Errorf("tag is '%s', expected: '%s'", is, wanted)
		must(err)
	}
}

func extractAttrByName(attrs []xml.Attr, attrName string) (*xml.Attr, []xml.Attr) {
	var rest []xml.Attr
	var res *xml.Attr
	for i, a := range attrs {
		if a.Name.Local == attrName {
			res = &attrs[i]
		} else {
			rest = append(rest, a)
		}
	}
	return res, rest
}

func mustExtractIntAttr(attrs []xml.Attr, attrName string, n *XMLNode) (int, []xml.Attr) {
	attr, rest := extractAttrByName(attrs, attrName)
	panicIf(attr == nil, "missing %s attribute in '%s'", attrName, n)
	i, err := strconv.Atoi(attr.Value)
	must(err)
	return i, rest
}

func mustExtractStringAttr(attrs []xml.Attr, attrName string, n *XMLNode) (string, []xml.Attr) {
	attr, rest := extractAttrByName(attrs, attrName)
	panicIf(attr == nil, "missing %s attribute in '%s'", attrName, n)
	return attr.Value, rest
}

func extractStringAttr(attrs []xml.Attr, name string, defValue string) (string, []xml.Attr) {
	attr, attrs := extractAttrByName(attrs, name)
	if attr == nil {
		return defValue, attrs
	}
	return attr.Value, attrs
}

func extractIntAttr(attrs []xml.Attr, name string, defValue int) (int, []xml.Attr) {
	attr, attrs := extractAttrByName(attrs, name)
	if attr == nil {
		return defValue, attrs
	}
	n, err := strconv.Atoi(attr.Value)
	must(err)
	return n, attrs
}

func extractBoolAttr(attrs []xml.Attr, name string, defValue bool) (bool, []xml.Attr) {
	attr, attrs := extractAttrByName(attrs, name)
	if attr == nil {
		return defValue, attrs
	}
	v := mustParseBool(attr.Value)
	return v, attrs
}
