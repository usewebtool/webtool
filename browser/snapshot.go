package browser

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-rod/rod/lib/proto"
)

// maxSnapshotDepth limits AX tree recursion to avoid runaway nesting.
const maxSnapshotDepth = 10

// maxTextLength is the maximum rune length for any text content in snapshots.
// All text (names, summaries, labels) is truncated to this length. If the LLM
// needs full text, it calls "webtool extract <backendNodeId>".
const maxTextLength = 160

// SnapshotMode controls the verbosity of the snapshot output.
type SnapshotMode int

const (
	// ModeDefault shows interactive elements, structural grouping, headings,
	// labels, status/alert, and content-container summaries.
	ModeDefault SnapshotMode = iota
	// ModeInteractive shows only interactive elements and structural grouping.
	ModeInteractive
	// ModeAll shows everything in default plus text-bearing content
	// (paragraphs, StaticText, blockquotes, code).
	ModeAll
)

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

	return formatSnapshot(info.URL, info.Title, result.Nodes, ModeDefault), nil
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

// contentContainerRoles are structural containers that represent one repeated
// content unit (inbox row, search result, product card). When unnamed, these
// get a synthetic summary from descendant text.
// Future candidates: row, cell, gridcell, columnheader, rowheader (deferred — table handling is separate).
var contentContainerRoles = map[string]bool{
	"listitem": true,
	"article":  true,
}

// classify determines how a node should be rendered based on its AX role and
// the current snapshot mode. Interactive roles are always shown. Info roles
// (headings, status messages, text content) are progressively included as the
// mode moves from interactive → default → all.
func classify(role string, hasName bool, mode SnapshotMode) nodeKind {
	if interactiveRoles[role] {
		return kindInteractive
	}

	// Info roles vary by mode: interactive shows none, default adds
	// headings/labels/images/status/alert, all adds text content on top.
	switch mode {
	case ModeInteractive:
		// Strip all info nodes — only interactive elements and structure.
	case ModeDefault:
		if role == "heading" || role == "LabelText" {
			return kindInfo
		}
		if role == "img" && hasName {
			return kindInfo
		}
		if role == "status" || role == "alert" {
			return kindInfo
		}
	case ModeAll:
		if role == "heading" || role == "LabelText" {
			return kindInfo
		}
		if role == "img" && hasName {
			return kindInfo
		}
		if role == "status" || role == "alert" {
			return kindInfo
		}
		// Text-bearing content only in all mode.
		if role == "paragraph" || role == "blockquote" || role == "code" || role == "StaticText" {
			return kindInfo
		}
	}

	if structuralRoles[role] {
		// Unnamed region/group/section are too generic to be useful — collapse them
		// so their children bubble up without extra nesting.
		if (role == "region" || role == "group" || role == "section") && !hasName {
			return kindCollapse
		}
		return kindStructural
	}
	return kindCollapse
}

// formatSnapshot builds the text snapshot from raw AX tree nodes.
func formatSnapshot(url, title string, nodes []*proto.AccessibilityAXNode, mode SnapshotMode) string {
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

	walkTree(&buf, rootID, nodeMap, childMap, 0, mode)

	return buf.String()
}

// walkTree recursively walks the AX tree, writing snapshot lines for relevant
// nodes. Returns true if the subtree contains "retainable" content — in default
// and interactive modes that means interactive elements; in all mode it also
// includes info nodes. Structural containers are only emitted when at least one
// retainable descendant exists (prevents empty nav/form/list noise).
func walkTree(
	buf *strings.Builder,
	nodeID proto.AccessibilityAXNodeID,
	nodeMap map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode,
	childMap map[proto.AccessibilityAXNodeID][]proto.AccessibilityAXNodeID,
	depth int,
	mode SnapshotMode,
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
			if walkTree(buf, childID, nodeMap, childMap, depth, mode) {
				hasInteractive = true
			}
		}
		return hasInteractive
	}

	role := axStr(node.Role)
	name := axStr(node.Name)
	kind := classify(role, name != "", mode)

	switch kind {
	case kindStructural:
		// Write children to a temp buffer first. We only emit the container
		// line if children produced retainable content — otherwise the
		// container is silently pruned from output.
		var childBuf strings.Builder
		hasInteractive := false
		for _, childID := range childMap[nodeID] {
			if walkTree(&childBuf, childID, nodeMap, childMap, depth+1, mode) {
				hasInteractive = true
			}
		}
		if !hasInteractive {
			return false
		}
		// For unnamed content containers (listitem, article), compute a
		// synthetic summary from non-interactive descendant text.
		if name == "" && contentContainerRoles[role] && mode != ModeInteractive {
			name = collectContainerText(nodeID, nodeMap, childMap)
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
		name = truncateText(name)
		formatNode(buf, node, role, name, depth)
		if kind == kindInteractive {
			return true
		}
		// Status and alert are important signals (error messages, confirmations)
		// that must retain parent containers in all modes — losing "Invalid
		// password" because it's inside a banner wrapper is a real problem.
		if role == "status" || role == "alert" {
			return true
		}
		// Other info nodes (headings, labels, images) don't retain parents
		// in default/interactive modes — a nav with only headings is pruned.
		// In all mode, info nodes DO retain parents so text-only containers
		// like a nav with paragraphs are shown.
		return mode == ModeAll

	default: // kindCollapse
		hasInteractive := false
		for _, childID := range childMap[nodeID] {
			if walkTree(buf, childID, nodeMap, childMap, depth, mode) {
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
	if role == "StaticText" {
		roleStr = "text"
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

// collectContainerText builds a synthetic summary for content containers
// (listitem, article) by gathering non-interactive text from descendants.
//
// This gives the LLM enough context to identify repeated content units (inbox
// rows, search results, product cards) without showing every StaticText node.
// For example, a Gmail listitem might produce "John Doe | Mar 10" while its
// interactive children (checkbox, subject link) are shown as separate lines.
//
// Key rules:
//   - Interactive children are skipped entirely (not recursed into), since
//     they already appear as their own lines in the snapshot.
//   - StaticText fragments that are whitespace-only or single-character are
//     skipped (decorative bullets, separators, etc.).
//   - Non-interactive, non-StaticText children are recursed into, picking up
//     text from formatting wrappers (spans, strongs, etc.).
//   - Fragments are joined with " | " for scannable readability.
//   - Result is truncated at maxTextLength.
func collectContainerText(
	nodeID proto.AccessibilityAXNodeID,
	nodeMap map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode,
	childMap map[proto.AccessibilityAXNodeID][]proto.AccessibilityAXNodeID,
) string {
	var fragments []string
	collectContainerTextWalk(nodeID, nodeMap, childMap, &fragments)

	joined := strings.Join(fragments, " | ")
	return truncateText(joined)
}

// collectContainerTextWalk is the recursive helper for collectContainerText.
// It appends non-interactive text fragments to the provided slice.
func collectContainerTextWalk(
	nodeID proto.AccessibilityAXNodeID,
	nodeMap map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode,
	childMap map[proto.AccessibilityAXNodeID][]proto.AccessibilityAXNodeID,
	fragments *[]string,
) {
	for _, childID := range childMap[nodeID] {
		child, ok := nodeMap[childID]
		if !ok {
			continue
		}
		role := axStr(child.Role)

		// Skip interactive elements entirely — their text is already shown
		// as separate child lines in the snapshot output.
		if interactiveRoles[role] {
			continue
		}

		if role == "StaticText" {
			text := strings.TrimSpace(axStr(child.Name))
			// Skip whitespace-only and single-char decorative text
			// (bullets, separators, icons, etc.).
			if len([]rune(text)) > 1 {
				*fragments = append(*fragments, text)
			}
			continue
		}

		// Recurse into non-interactive wrappers (generic divs, spans,
		// formatting elements) to pick up their StaticText descendants.
		collectContainerTextWalk(childID, nodeMap, childMap, fragments)
	}
}

// truncateText truncates s to maxTextLength runes, appending "..." if truncated.
func truncateText(s string) string {
	runes := []rune(s)
	if len(runes) <= maxTextLength {
		return s
	}
	return string(runes[:maxTextLength-3]) + "..."
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
