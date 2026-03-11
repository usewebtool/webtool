package browser

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-rod/rod/lib/proto"
)

const maxSnapshotDepth = 10

// Snapshot returns a token-efficient text representation of the current page's
// interactive elements, structured by accessibility landmarks and forms.
func (b *Browser) Snapshot(ctx context.Context) (string, error) {
	if err := b.Connect(); err != nil {
		return "", err
	}

	page, err := b.activePage()
	if err != nil {
		return "", err
	}
	page = page.Context(ctx)

	info, err := page.Info()
	if err != nil {
		return "", fmt.Errorf("getting page info: %w", err)
	}

	result, err := proto.AccessibilityGetFullAXTree{}.Call(page)
	if err != nil {
		return "", fmt.Errorf("getting accessibility tree: %w", err)
	}

	return formatSnapshot(info.URL, info.Title, result.Nodes), nil
}

// nodeKind classifies how a node should be handled during tree walking.
type nodeKind int

const (
	kindCollapse    nodeKind = iota // don't show, recurse children at same depth
	kindStructural                  // show + recurse children at depth+1
	kindInteractive                 // show as leaf (no recursion)
	kindInfo                        // show as leaf (no recursion)
)

// structuralRoles are container roles that provide page structure (landmarks, forms, lists, etc.).
var structuralRoles = map[string]bool{
	"form":          true,
	"navigation":    true,
	"main":          true,
	"dialog":        true,
	"alertdialog":   true,
	"banner":        true,
	"contentinfo":   true,
	"complementary": true,
	"search":        true,
	"toolbar":       true,
	"menu":          true,
	"menubar":       true,
	"list":          true,
	"listbox":       true,
	"listitem":      true,
	"table":         true,
	"tree":          true,
	"tablist":       true,
	"region":        true,
	"group":         true,
	"article":       true,
	"section":       true,
}

// interactiveRoles are elements the user can act on (click, type, select, etc.).
var interactiveRoles = map[string]bool{
	"button":            true,
	"link":              true,
	"textbox":           true,
	"searchbox":         true,
	"combobox":          true,
	"checkbox":          true,
	"radio":             true,
	"slider":            true,
	"spinbutton":        true,
	"switch":            true,
	"menuitem":          true,
	"menuitemcheckbox":  true,
	"menuitemradio":     true,
	"option":            true,
	"tab":               true,
	"treeitem":          true,
	"textarea":          true,
	"select":            true,
	"ProgressIndicator": true,
}

func classify(role string, hasName bool) nodeKind {
	if interactiveRoles[role] {
		return kindInteractive
	}
	if role == "heading" || role == "LabelText" {
		return kindInfo
	}
	if role == "img" && hasName {
		return kindInfo
	}
	if structuralRoles[role] {
		// region, group, and section only shown if named
		if (role == "region" || role == "group" || role == "section") && !hasName {
			return kindCollapse
		}
		return kindStructural
	}
	return kindCollapse
}

// formatSnapshot builds the text snapshot from raw AX tree nodes.
func formatSnapshot(url, title string, nodes []*proto.AccessibilityAXNode) string {
	if len(nodes) == 0 {
		return fmt.Sprintf("[url] %s\n[title] %s\n", url, title)
	}

	// Build lookup maps.
	nodeMap := make(map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode, len(nodes))
	childMap := make(map[proto.AccessibilityAXNodeID][]proto.AccessibilityAXNodeID, len(nodes))
	for _, n := range nodes {
		nodeMap[n.NodeID] = n
		if len(n.ChildIDs) > 0 {
			childMap[n.NodeID] = n.ChildIDs
		}
	}

	// Find root (first node without a parent).
	var rootID proto.AccessibilityAXNodeID
	for _, n := range nodes {
		if n.ParentID == "" {
			rootID = n.NodeID
			break
		}
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "[url] %s\n[title] %s\n", url, title)

	walkTree(&buf, rootID, nodeMap, childMap, 0)

	return buf.String()
}

// walkTree recursively walks the AX tree, writing interactive elements with
// structural indentation. Returns true if any interactive elements were written.
func walkTree(
	buf *strings.Builder,
	nodeID proto.AccessibilityAXNodeID,
	nodeMap map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode,
	childMap map[proto.AccessibilityAXNodeID][]proto.AccessibilityAXNodeID,
	depth int,
) bool {
	if depth > maxSnapshotDepth {
		return false
	}

	node, ok := nodeMap[nodeID]
	if !ok {
		return false
	}

	// Ignored nodes: don't show, but recurse children at same depth.
	if node.Ignored {
		hasInteractive := false
		for _, childID := range childMap[nodeID] {
			if walkTree(buf, childID, nodeMap, childMap, depth) {
				hasInteractive = true
			}
		}
		return hasInteractive
	}

	role := axStr(node.Role)
	name := axStr(node.Name)
	kind := classify(role, name != "")

	switch kind {
	case kindStructural:
		// Write children to a temp buffer; only emit the container if
		// at least one interactive descendant was found.
		var childBuf strings.Builder
		hasInteractive := false
		for _, childID := range childMap[nodeID] {
			if walkTree(&childBuf, childID, nodeMap, childMap, depth+1) {
				hasInteractive = true
			}
		}
		if !hasInteractive {
			return false
		}
		formatNode(buf, node, role, name, depth)
		buf.WriteString(childBuf.String())
		return true

	case kindInteractive, kindInfo:
		// If the node has no name, try to resolve it from StaticText descendants.
		if name == "" {
			name = collectStaticText(nodeID, nodeMap, childMap)
		}
		// Skip info nodes that have no text to show.
		if kind == kindInfo && name == "" {
			return false
		}
		formatNode(buf, node, role, name, depth)
		return kind == kindInteractive

	default: // kindCollapse
		hasInteractive := false
		for _, childID := range childMap[nodeID] {
			if walkTree(buf, childID, nodeMap, childMap, depth) {
				hasInteractive = true
			}
		}
		return hasInteractive
	}
}

// formatNode writes a single snapshot line for a node.
func formatNode(buf *strings.Builder, node *proto.AccessibilityAXNode, role, name string, depth int) {
	indent := strings.Repeat("  ", depth)

	// Map AX role names to more readable display names.
	roleStr := role
	if role == "LabelText" {
		roleStr = "label"
	}
	if role == "heading" {
		if lvl := nodeProperty(node, "level"); lvl != "" {
			roleStr = fmt.Sprintf("heading[%s]", lvl)
		}
	}

	// Start the line: indent [backendNodeId] role
	if node.BackendDOMNodeID != 0 {
		fmt.Fprintf(buf, "%s[%d] %s", indent, node.BackendDOMNodeID, roleStr)
	} else {
		fmt.Fprintf(buf, "%s%s", indent, roleStr)
	}

	// Name
	if name != "" {
		fmt.Fprintf(buf, " %q", name)
	}

	// Value (for inputs with current content)
	if val := axStr(node.Value); val != "" {
		fmt.Fprintf(buf, " value=%q", val)
	}

	// URL for links (stripped of query params to save tokens).
	if linkURL := nodeProperty(node, "url"); linkURL != "" {
		fmt.Fprintf(buf, " url=%q", stripQueryString(linkURL))
	}

	// State flags
	if nodePropertyBool(node, "focused") {
		buf.WriteString(" focused")
	}
	if nodePropertyBool(node, "checked") {
		buf.WriteString(" checked")
	}
	if nodePropertyBool(node, "disabled") {
		buf.WriteString(" disabled")
	}
	if nodePropertyBool(node, "readonly") {
		buf.WriteString(" readonly")
	}
	if nodePropertyBool(node, "required") {
		buf.WriteString(" required")
	}
	if nodePropertyBool(node, "selected") {
		buf.WriteString(" selected")
	}
	if nodePropertyBool(node, "expanded") {
		buf.WriteString(" expanded")
	}

	buf.WriteByte('\n')
}

// collectStaticText recursively collects text from StaticText descendants.
// Used to resolve the display text for nodes like LabelText where Chrome puts
// the visible text in StaticText children rather than the node's own name.
func collectStaticText(
	nodeID proto.AccessibilityAXNodeID,
	nodeMap map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode,
	childMap map[proto.AccessibilityAXNodeID][]proto.AccessibilityAXNodeID,
) string {
	var parts []string
	for _, childID := range childMap[nodeID] {
		child, ok := nodeMap[childID]
		if !ok {
			continue
		}
		if axStr(child.Role) == "StaticText" {
			if t := axStr(child.Name); t != "" {
				parts = append(parts, t)
			}
		} else {
			// Recurse into formatting wrappers (e.g. <strong>, <em>).
			if t := collectStaticText(childID, nodeMap, childMap); t != "" {
				parts = append(parts, t)
			}
		}
	}
	return strings.Join(parts, "")
}

// axStr extracts the string value from an AXValue, or "" if nil/empty.
func axStr(v *proto.AccessibilityAXValue) string {
	if v == nil {
		return ""
	}
	return v.Value.Str()
}

// nodeProperty returns the string value of a named property, or "".
func nodeProperty(node *proto.AccessibilityAXNode, name proto.AccessibilityAXPropertyName) string {
	for _, p := range node.Properties {
		if p.Name == name && p.Value != nil {
			return p.Value.Value.Str()
		}
	}
	return ""
}

// nodePropertyBool returns true if a named boolean property is "true".
func nodePropertyBool(node *proto.AccessibilityAXNode, name proto.AccessibilityAXPropertyName) bool {
	return nodeProperty(node, name) == "true"
}

// stripQueryString removes the query string and fragment from a URL.
func stripQueryString(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
