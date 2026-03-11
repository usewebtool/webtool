package browser

import (
	"strings"
	"testing"

	"github.com/go-rod/rod/lib/proto"
	"github.com/ysmood/gson"
)

// axVal creates an AccessibilityAXValue with a string value.
func axVal(s string) *proto.AccessibilityAXValue {
	return &proto.AccessibilityAXValue{
		Type:  "string",
		Value: gson.New(s),
	}
}

// axBoolVal creates an AccessibilityAXValue with a boolean value.
func axBoolVal(b bool) *proto.AccessibilityAXValue {
	if b {
		return &proto.AccessibilityAXValue{Type: "boolean", Value: gson.New("true")}
	}
	return &proto.AccessibilityAXValue{Type: "boolean", Value: gson.New("false")}
}

// prop creates an AXProperty.
func prop(name proto.AccessibilityAXPropertyName, val *proto.AccessibilityAXValue) *proto.AccessibilityAXProperty {
	return &proto.AccessibilityAXProperty{Name: name, Value: val}
}

func TestFormatSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		title    string
		nodes    []*proto.AccessibilityAXNode
		contains []string // lines that must appear in output
		excludes []string // lines that must NOT appear in output
	}{
		{
			name:  "empty tree",
			url:   "https://example.com",
			title: "Example",
			nodes: nil,
			contains: []string{
				"[url] https://example.com",
				"[title] Example",
			},
		},
		{
			name:  "form with inputs",
			url:   "https://example.com/login",
			title: "Login",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"main1"}},
				{NodeID: "main1", ParentID: "root", Role: axVal("main"), BackendDOMNodeID: 1, ChildIDs: []proto.AccessibilityAXNodeID{"form1"}},
				{NodeID: "form1", ParentID: "main1", Role: axVal("form"), Name: axVal("Login"), BackendDOMNodeID: 10, ChildIDs: []proto.AccessibilityAXNodeID{"email", "pass", "submit"}},
				{NodeID: "email", ParentID: "form1", Role: axVal("textbox"), Name: axVal("Email"), BackendDOMNodeID: 11},
				{NodeID: "pass", ParentID: "form1", Role: axVal("textbox"), Name: axVal("Password"), BackendDOMNodeID: 12},
				{NodeID: "submit", ParentID: "form1", Role: axVal("button"), Name: axVal("Sign in"), BackendDOMNodeID: 13},
			},
			contains: []string{
				"[1] main",
				"  [10] form \"Login\"",
				"    [11] textbox \"Email\"",
				"    [12] textbox \"Password\"",
				"    [13] button \"Sign in\"",
			},
		},
		{
			name:  "unnamed form still shown",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"form1"}},
				{NodeID: "form1", ParentID: "root", Role: axVal("form"), BackendDOMNodeID: 20, ChildIDs: []proto.AccessibilityAXNodeID{"input1"}},
				{NodeID: "input1", ParentID: "form1", Role: axVal("textbox"), Name: axVal("Search"), BackendDOMNodeID: 21},
			},
			contains: []string{
				"[20] form",
				"  [21] textbox \"Search\"",
			},
		},
		{
			name:  "generic containers collapsed",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"div1"}},
				{NodeID: "div1", ParentID: "root", Role: axVal("generic"), BackendDOMNodeID: 30, ChildIDs: []proto.AccessibilityAXNodeID{"div2"}},
				{NodeID: "div2", ParentID: "div1", Role: axVal("generic"), BackendDOMNodeID: 31, ChildIDs: []proto.AccessibilityAXNodeID{"btn1"}},
				{NodeID: "btn1", ParentID: "div2", Role: axVal("button"), Name: axVal("Click me"), BackendDOMNodeID: 32},
			},
			contains: []string{
				"[32] button \"Click me\"",
			},
			excludes: []string{
				"generic",
				"[30]",
				"[31]",
			},
		},
		{
			name:  "structural container pruned when no interactive descendants",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"nav1", "btn1"}},
				{NodeID: "nav1", ParentID: "root", Role: axVal("navigation"), BackendDOMNodeID: 40, ChildIDs: []proto.AccessibilityAXNodeID{"text1"}},
				{NodeID: "text1", ParentID: "nav1", Role: axVal("StaticText"), BackendDOMNodeID: 41},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("OK"), BackendDOMNodeID: 42},
			},
			contains: []string{
				"[42] button \"OK\"",
			},
			excludes: []string{
				"navigation",
				"[40]",
			},
		},
		{
			name:  "heading with level",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"h1", "link1"}},
				{NodeID: "h1", ParentID: "root", Role: axVal("heading"), Name: axVal("Welcome"),
					BackendDOMNodeID: 50,
					Properties:       []*proto.AccessibilityAXProperty{prop("level", axVal("1"))}},
				{NodeID: "link1", ParentID: "root", Role: axVal("link"), Name: axVal("About"), BackendDOMNodeID: 51},
			},
			contains: []string{
				`[50] heading[1] "Welcome"`,
				`[51] link "About"`,
			},
		},
		{
			name:  "input with value",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"input1"}},
				{NodeID: "input1", ParentID: "root", Role: axVal("textbox"), Name: axVal("Email"),
					BackendDOMNodeID: 60,
					Value:            axVal("user@example.com")},
			},
			contains: []string{
				`[60] textbox "Email" value="user@example.com"`,
			},
		},
		{
			name:  "checkbox checked and disabled button",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"cb1", "btn1"}},
				{NodeID: "cb1", ParentID: "root", Role: axVal("checkbox"), Name: axVal("Remember me"),
					BackendDOMNodeID: 70,
					Properties:       []*proto.AccessibilityAXProperty{prop("checked", axBoolVal(true))}},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("Submit"),
					BackendDOMNodeID: 71,
					Properties:       []*proto.AccessibilityAXProperty{prop("disabled", axBoolVal(true))}},
			},
			contains: []string{
				`[70] checkbox "Remember me" checked`,
				`[71] button "Submit" disabled`,
			},
		},
		{
			name:  "ignored node with non-ignored children",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"ignored1"}},
				{NodeID: "ignored1", ParentID: "root", Ignored: true, BackendDOMNodeID: 80, ChildIDs: []proto.AccessibilityAXNodeID{"btn1"}},
				{NodeID: "btn1", ParentID: "ignored1", Role: axVal("button"), Name: axVal("Visible"), BackendDOMNodeID: 81},
			},
			contains: []string{
				`[81] button "Visible"`,
			},
			excludes: []string{
				"[80]",
			},
		},
		{
			name:  "unnamed region collapsed but named region shown",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"r1", "r2"}},
				{NodeID: "r1", ParentID: "root", Role: axVal("region"), BackendDOMNodeID: 90, ChildIDs: []proto.AccessibilityAXNodeID{"btn1"}},
				{NodeID: "btn1", ParentID: "r1", Role: axVal("button"), Name: axVal("A"), BackendDOMNodeID: 91},
				{NodeID: "r2", ParentID: "root", Role: axVal("region"), Name: axVal("Sidebar"), BackendDOMNodeID: 92, ChildIDs: []proto.AccessibilityAXNodeID{"btn2"}},
				{NodeID: "btn2", ParentID: "r2", Role: axVal("button"), Name: axVal("B"), BackendDOMNodeID: 93},
			},
			contains: []string{
				`[91] button "A"`,
				`[92] region "Sidebar"`,
				`  [93] button "B"`,
			},
			excludes: []string{
				"[90]", // unnamed region not shown
			},
		},
		{
			name:  "link with url stripped of query params",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"link1", "link2"}},
				{NodeID: "link1", ParentID: "root", Role: axVal("link"), Name: axVal("Search"), BackendDOMNodeID: 110,
					Properties: []*proto.AccessibilityAXProperty{prop("url", axVal("https://example.com/search?q=foo&utm_source=bar"))}},
				{NodeID: "link2", ParentID: "root", Role: axVal("link"), Name: axVal("About"), BackendDOMNodeID: 111,
					Properties: []*proto.AccessibilityAXProperty{prop("url", axVal("https://example.com/about"))}},
			},
			contains: []string{
				`[110] link "Search" url="https://example.com/search"`,
				`[111] link "About" url="https://example.com/about"`,
			},
			excludes: []string{
				"q=foo",
				"utm_source",
			},
		},
		{
			name:  "selected tab and readonly input",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"tab1", "tab2", "input1"}},
				{NodeID: "tab1", ParentID: "root", Role: axVal("tab"), Name: axVal("General"), BackendDOMNodeID: 120,
					Properties: []*proto.AccessibilityAXProperty{prop("selected", axBoolVal(true))}},
				{NodeID: "tab2", ParentID: "root", Role: axVal("tab"), Name: axVal("Advanced"), BackendDOMNodeID: 121},
				{NodeID: "input1", ParentID: "root", Role: axVal("textbox"), Name: axVal("ID"), BackendDOMNodeID: 122,
					Value:      axVal("abc-123"),
					Properties: []*proto.AccessibilityAXProperty{prop("readonly", axBoolVal(true))}},
			},
			contains: []string{
				`[120] tab "General" selected`,
				`[121] tab "Advanced"`,
				`[122] textbox "ID" value="abc-123" readonly`,
			},
		},
		{
			name:  "focused element",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"input1"}},
				{NodeID: "input1", ParentID: "root", Role: axVal("textbox"), Name: axVal("Search"), BackendDOMNodeID: 130,
					Properties: []*proto.AccessibilityAXProperty{prop("focused", axBoolVal(true))}},
			},
			contains: []string{
				`[130] textbox "Search" focused`,
			},
		},
		{
			name:  "listitem shown when containing interactive elements",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"list1"}},
				{NodeID: "list1", ParentID: "root", Role: axVal("list"), BackendDOMNodeID: 200, ChildIDs: []proto.AccessibilityAXNodeID{"li1", "li2"}},
				{NodeID: "li1", ParentID: "list1", Role: axVal("listitem"), BackendDOMNodeID: 201, ChildIDs: []proto.AccessibilityAXNodeID{"link1"}},
				{NodeID: "link1", ParentID: "li1", Role: axVal("link"), Name: axVal("Home"), BackendDOMNodeID: 202},
				{NodeID: "li2", ParentID: "list1", Role: axVal("listitem"), BackendDOMNodeID: 203, ChildIDs: []proto.AccessibilityAXNodeID{"link2"}},
				{NodeID: "link2", ParentID: "li2", Role: axVal("link"), Name: axVal("About"), BackendDOMNodeID: 204},
			},
			contains: []string{
				"[200] list",
				"  [201] listitem",
				"    [202] link \"Home\"",
				"  [203] listitem",
				"    [204] link \"About\"",
			},
		},
		{
			name:  "listitem pruned when no interactive descendants",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"list1"}},
				{NodeID: "list1", ParentID: "root", Role: axVal("list"), BackendDOMNodeID: 210, ChildIDs: []proto.AccessibilityAXNodeID{"li1"}},
				{NodeID: "li1", ParentID: "list1", Role: axVal("listitem"), BackendDOMNodeID: 211, ChildIDs: []proto.AccessibilityAXNodeID{"text1"}},
				{NodeID: "text1", ParentID: "li1", Role: axVal("StaticText"), Name: axVal("Just text"), BackendDOMNodeID: 212},
			},
			excludes: []string{
				"listitem",
				"[210]",
			},
		},
		{
			name:  "article shown with interactive descendants",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"art1"}},
				{NodeID: "art1", ParentID: "root", Role: axVal("article"), Name: axVal("Blog Post"), BackendDOMNodeID: 220, ChildIDs: []proto.AccessibilityAXNodeID{"link1"}},
				{NodeID: "link1", ParentID: "art1", Role: axVal("link"), Name: axVal("Read more"), BackendDOMNodeID: 221},
			},
			contains: []string{
				`[220] article "Blog Post"`,
				`  [221] link "Read more"`,
			},
		},
		{
			name:  "named section shown, unnamed section collapsed",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"s1", "s2"}},
				{NodeID: "s1", ParentID: "root", Role: axVal("section"), Name: axVal("Sidebar"), BackendDOMNodeID: 230, ChildIDs: []proto.AccessibilityAXNodeID{"btn1"}},
				{NodeID: "btn1", ParentID: "s1", Role: axVal("button"), Name: axVal("Toggle"), BackendDOMNodeID: 231},
				{NodeID: "s2", ParentID: "root", Role: axVal("section"), BackendDOMNodeID: 232, ChildIDs: []proto.AccessibilityAXNodeID{"btn2"}},
				{NodeID: "btn2", ParentID: "s2", Role: axVal("button"), Name: axVal("OK"), BackendDOMNodeID: 233},
			},
			contains: []string{
				`[230] section "Sidebar"`,
				`  [231] button "Toggle"`,
				`[233] button "OK"`,
			},
			excludes: []string{
				"[232]", // unnamed section not shown
			},
		},
		{
			name:  "img with alt text shown, without alt text hidden",
			url:   "https://example.com",
			title: "Test",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"img1", "img2", "btn1"}},
				{NodeID: "img1", ParentID: "root", Role: axVal("img"), Name: axVal("Logo"), BackendDOMNodeID: 100},
				{NodeID: "img2", ParentID: "root", Role: axVal("img"), BackendDOMNodeID: 101},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("Go"), BackendDOMNodeID: 102},
			},
			contains: []string{
				`[100] img "Logo"`,
				`[102] button "Go"`,
			},
			excludes: []string{
				"[101]", // img without alt text
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSnapshot(tt.url, tt.title, tt.nodes)

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("output missing expected line %q\n\ngot:\n%s", want, got)
				}
			}
			for _, reject := range tt.excludes {
				if strings.Contains(got, reject) {
					t.Errorf("output should not contain %q\n\ngot:\n%s", reject, got)
				}
			}
		})
	}
}
