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
			got := formatSnapshot(tt.url, tt.title, tt.nodes, ModeDefault)

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

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"short", "hello", "hello"},
		{"exact limit", strings.Repeat("a", 160), strings.Repeat("a", 160)},
		{"over limit", strings.Repeat("a", 170), strings.Repeat("a", 157) + "..."},
		{"empty", "", ""},
		{"multibyte runes", strings.Repeat("日", 165), strings.Repeat("日", 157) + "..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, maxTextLength)
			if got != tt.want {
				t.Errorf("truncate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		hasName bool
		mode    SnapshotMode
		want    nodeKind
	}{
		// Interactive roles are always interactive regardless of mode.
		{"button default", "button", true, ModeDefault, kindInteractive},
		{"button interactive", "button", true, ModeInteractive, kindInteractive},
		{"button all", "button", true, ModeAll, kindInteractive},
		{"link", "link", true, ModeDefault, kindInteractive},

		// Heading: info in default/all, collapsed in interactive.
		{"heading default", "heading", true, ModeDefault, kindInfo},
		{"heading interactive", "heading", true, ModeInteractive, kindCollapse},
		{"heading all", "heading", true, ModeAll, kindInfo},

		// LabelText: info in default/all, collapsed in interactive.
		{"label default", "LabelText", true, ModeDefault, kindInfo},
		{"label interactive", "LabelText", true, ModeInteractive, kindCollapse},

		// img with name: info in default/all, collapsed in interactive.
		{"img named default", "img", true, ModeDefault, kindInfo},
		{"img named interactive", "img", true, ModeInteractive, kindCollapse},
		{"img named all", "img", true, ModeAll, kindInfo},
		// img without name: always collapsed.
		{"img unnamed", "img", false, ModeDefault, kindCollapse},

		// status/alert: info in default/all, collapsed in interactive.
		{"status default", "status", false, ModeDefault, kindInfo},
		{"status interactive", "status", false, ModeInteractive, kindCollapse},
		{"alert default", "alert", false, ModeDefault, kindInfo},
		{"alert all", "alert", false, ModeAll, kindInfo},

		// paragraph/blockquote/code/StaticText: info only in all mode.
		{"paragraph default", "paragraph", false, ModeDefault, kindCollapse},
		{"paragraph all", "paragraph", false, ModeAll, kindInfo},
		{"blockquote all", "blockquote", false, ModeAll, kindInfo},
		{"code all", "code", false, ModeAll, kindInfo},
		{"StaticText all", "StaticText", false, ModeAll, kindInfo},
		{"StaticText default", "StaticText", false, ModeDefault, kindCollapse},

		// Structural roles.
		{"form", "form", true, ModeDefault, kindStructural},
		{"navigation", "navigation", false, ModeDefault, kindStructural},
		{"listitem", "listitem", false, ModeDefault, kindStructural},
		{"article", "article", false, ModeDefault, kindStructural},

		// Unnamed region/group/section collapse.
		{"unnamed region", "region", false, ModeDefault, kindCollapse},
		{"named region", "region", true, ModeDefault, kindStructural},
		{"unnamed group", "group", false, ModeDefault, kindCollapse},
		{"unnamed section", "section", false, ModeDefault, kindCollapse},
		{"named section", "section", true, ModeDefault, kindStructural},

		// Unknown roles collapse.
		{"generic", "generic", false, ModeDefault, kindCollapse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classify(tt.role, tt.hasName, tt.mode)
			if got != tt.want {
				t.Errorf("classify(%q, %v, %v) = %v, want %v", tt.role, tt.hasName, tt.mode, got, tt.want)
			}
		})
	}
}

func TestFormatSnapshotModes(t *testing.T) {
	tests := []struct {
		name     string
		mode     SnapshotMode
		nodes    []*proto.AccessibilityAXNode
		contains []string
		excludes []string
	}{
		{
			name: "interactive mode strips headings, images, and labels",
			mode: ModeInteractive,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"h1", "img1", "lbl1", "btn1"}},
				{NodeID: "h1", ParentID: "root", Role: axVal("heading"), Name: axVal("Welcome"), BackendDOMNodeID: 50,
					Properties: []*proto.AccessibilityAXProperty{prop("level", axVal("1"))}},
				{NodeID: "img1", ParentID: "root", Role: axVal("img"), Name: axVal("Logo"), BackendDOMNodeID: 51},
				{NodeID: "lbl1", ParentID: "root", Role: axVal("LabelText"), Name: axVal("Email"), BackendDOMNodeID: 53},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("Go"), BackendDOMNodeID: 52},
			},
			contains: []string{`[52] button "Go"`},
			excludes: []string{"heading", "Welcome", "img", "Logo", "label", "Email"},
		},
		{
			name: "default mode shows status and alert",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"s1", "a1", "btn1"}},
				{NodeID: "s1", ParentID: "root", Role: axVal("status"), Name: axVal("Item added"), BackendDOMNodeID: 60},
				{NodeID: "a1", ParentID: "root", Role: axVal("alert"), Name: axVal("Invalid password"), BackendDOMNodeID: 61},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("OK"), BackendDOMNodeID: 62},
			},
			contains: []string{
				`[60] status "Item added"`,
				`[61] alert "Invalid password"`,
				`[62] button "OK"`,
			},
		},
		{
			name: "interactive mode strips status and alert",
			mode: ModeInteractive,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"s1", "btn1"}},
				{NodeID: "s1", ParentID: "root", Role: axVal("status"), Name: axVal("Item added"), BackendDOMNodeID: 60},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("OK"), BackendDOMNodeID: 62},
			},
			contains: []string{`[62] button "OK"`},
			excludes: []string{"status", "Item added"},
		},
		{
			name: "all mode shows paragraph and StaticText",
			mode: ModeAll,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"p1", "st1", "btn1"}},
				{NodeID: "p1", ParentID: "root", Role: axVal("paragraph"), Name: axVal("Some paragraph text"), BackendDOMNodeID: 70},
				{NodeID: "st1", ParentID: "root", Role: axVal("StaticText"), Name: axVal("Raw text"), BackendDOMNodeID: 71},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("OK"), BackendDOMNodeID: 72},
			},
			contains: []string{
				`[70] paragraph "Some paragraph text"`,
				`[71] text "Raw text"`,
				`[72] button "OK"`,
			},
		},
		{
			name: "all mode shows blockquote and code",
			mode: ModeAll,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"bq1", "code1", "btn1"}},
				{NodeID: "bq1", ParentID: "root", Role: axVal("blockquote"), Name: axVal("To be or not to be"), BackendDOMNodeID: 73},
				{NodeID: "code1", ParentID: "root", Role: axVal("code"), Name: axVal("fmt.Println(hello)"), BackendDOMNodeID: 74},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("OK"), BackendDOMNodeID: 75},
			},
			contains: []string{
				`[73] blockquote "To be or not to be"`,
				`[74] code "fmt.Println(hello)"`,
				`[75] button "OK"`,
			},
		},
		{
			name: "default mode hides paragraph and StaticText",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"p1", "btn1"}},
				{NodeID: "p1", ParentID: "root", Role: axVal("paragraph"), Name: axVal("Some text"), BackendDOMNodeID: 70},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("OK"), BackendDOMNodeID: 72},
			},
			contains: []string{`[72] button "OK"`},
			excludes: []string{"paragraph", "Some text"},
		},
		{
			name: "all mode retains structural container with text-only content",
			mode: ModeAll,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"nav1"}},
				{NodeID: "nav1", ParentID: "root", Role: axVal("navigation"), BackendDOMNodeID: 80, ChildIDs: []proto.AccessibilityAXNodeID{"p1"}},
				{NodeID: "p1", ParentID: "nav1", Role: axVal("paragraph"), Name: axVal("Breadcrumb text"), BackendDOMNodeID: 81},
			},
			contains: []string{
				"[80] navigation",
				`  [81] paragraph "Breadcrumb text"`,
			},
		},
		{
			name: "default mode prunes structural container with text-only content",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"nav1"}},
				{NodeID: "nav1", ParentID: "root", Role: axVal("navigation"), BackendDOMNodeID: 80, ChildIDs: []proto.AccessibilityAXNodeID{"st1"}},
				{NodeID: "st1", ParentID: "nav1", Role: axVal("StaticText"), Name: axVal("Just text"), BackendDOMNodeID: 81},
			},
			excludes: []string{"navigation", "[80]"},
		},
		{
			name: "status inside structural container retained in default mode",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"banner1"}},
				{NodeID: "banner1", ParentID: "root", Role: axVal("banner"), BackendDOMNodeID: 95, ChildIDs: []proto.AccessibilityAXNodeID{"s1"}},
				{NodeID: "s1", ParentID: "banner1", Role: axVal("status"), Name: axVal("Item added to cart"), BackendDOMNodeID: 96},
			},
			contains: []string{
				"[95] banner",
				`  [96] status "Item added to cart"`,
			},
		},
		{
			name: "alert inside structural container retained in default mode",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"region1"}},
				{NodeID: "region1", ParentID: "root", Role: axVal("region"), Name: axVal("Notifications"), BackendDOMNodeID: 97, ChildIDs: []proto.AccessibilityAXNodeID{"a1"}},
				{NodeID: "a1", ParentID: "region1", Role: axVal("alert"), Name: axVal("Invalid password"), BackendDOMNodeID: 98},
			},
			contains: []string{
				`[97] region "Notifications"`,
				`  [98] alert "Invalid password"`,
			},
		},
		{
			name: "text truncated at maxTextLength",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"btn1"}},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal(strings.Repeat("x", 200)), BackendDOMNodeID: 90},
			},
			contains: []string{strings.Repeat("x", 157) + "..."},
			excludes: []string{strings.Repeat("x", 158)},
		},
		{
			name: "invalid state flag shown on form control",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"input1"}},
				{NodeID: "input1", ParentID: "root", Role: axVal("textbox"), Name: axVal("Email"), BackendDOMNodeID: 400,
					Properties: []*proto.AccessibilityAXProperty{prop("invalid", axBoolVal(true))}},
			},
			contains: []string{`[400] textbox "Email" invalid`},
		},
		{
			name: "collapsed shown when expanded is false",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"btn1"}},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("Menu"), BackendDOMNodeID: 401,
					Properties: []*proto.AccessibilityAXProperty{prop("expanded", axBoolVal(false))}},
			},
			contains: []string{`[401] button "Menu" collapsed`},
			excludes: []string{"expanded"},
		},
		{
			name: "expanded shown when expanded is true",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"btn1"}},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("Menu"), BackendDOMNodeID: 402,
					Properties: []*proto.AccessibilityAXProperty{prop("expanded", axBoolVal(true))}},
			},
			contains: []string{`[402] button "Menu" expanded`},
			excludes: []string{"collapsed"},
		},
		{
			name: "no expanded or collapsed when property absent",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"btn1"}},
				{NodeID: "btn1", ParentID: "root", Role: axVal("button"), Name: axVal("Submit"), BackendDOMNodeID: 403},
			},
			contains: []string{`[403] button "Submit"`},
			excludes: []string{"expanded", "collapsed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSnapshot("https://example.com", "Test", tt.nodes, tt.mode)

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

func TestCollectStaticText(t *testing.T) {
	buildMaps := func(nodes []*proto.AccessibilityAXNode) (
		map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode,
		map[proto.AccessibilityAXNodeID][]proto.AccessibilityAXNodeID,
	) {
		nodeMap := make(map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode, len(nodes))
		childMap := make(map[proto.AccessibilityAXNodeID][]proto.AccessibilityAXNodeID, len(nodes))
		for _, n := range nodes {
			nodeMap[n.NodeID] = n
			if len(n.ChildIDs) > 0 {
				childMap[n.NodeID] = n.ChildIDs
			}
		}
		return nodeMap, childMap
	}

	tests := []struct {
		name  string
		nodes []*proto.AccessibilityAXNode
		root  proto.AccessibilityAXNodeID
		want  string
	}{
		{
			name: "collects direct StaticText children",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "label", Role: axVal("LabelText"), ChildIDs: []proto.AccessibilityAXNodeID{"t1", "t2"}},
				{NodeID: "t1", ParentID: "label", Role: axVal("StaticText"), Name: axVal("First ")},
				{NodeID: "t2", ParentID: "label", Role: axVal("StaticText"), Name: axVal("Name")},
			},
			root: "label",
			want: "First Name",
		},
		{
			name: "recurses into formatting wrappers",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "label", Role: axVal("LabelText"), ChildIDs: []proto.AccessibilityAXNodeID{"strong1"}},
				{NodeID: "strong1", ParentID: "label", Role: axVal("generic"), ChildIDs: []proto.AccessibilityAXNodeID{"t1"}},
				{NodeID: "t1", ParentID: "strong1", Role: axVal("StaticText"), Name: axVal("Bold text")},
			},
			root: "label",
			want: "Bold text",
		},
		{
			name: "skips empty StaticText",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "label", Role: axVal("LabelText"), ChildIDs: []proto.AccessibilityAXNodeID{"t1", "t2"}},
				{NodeID: "t1", ParentID: "label", Role: axVal("StaticText"), Name: axVal("")},
				{NodeID: "t2", ParentID: "label", Role: axVal("StaticText"), Name: axVal("Visible")},
			},
			root: "label",
			want: "Visible",
		},
		{
			name:  "returns empty for no children",
			nodes: []*proto.AccessibilityAXNode{{NodeID: "label", Role: axVal("LabelText")}},
			root:  "label",
			want:  "",
		},
		{
			name: "handles missing child node",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "label", Role: axVal("LabelText"), ChildIDs: []proto.AccessibilityAXNodeID{"missing", "t1"}},
				{NodeID: "t1", ParentID: "label", Role: axVal("StaticText"), Name: axVal("OK")},
			},
			root: "label",
			want: "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeMap, childMap := buildMaps(tt.nodes)
			got := collectStaticText(tt.root, nodeMap, childMap)
			if got != tt.want {
				t.Errorf("collectStaticText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCollectContainerText(t *testing.T) {
	buildMaps := func(nodes []*proto.AccessibilityAXNode) (
		map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode,
		map[proto.AccessibilityAXNodeID][]proto.AccessibilityAXNodeID,
	) {
		nodeMap := make(map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode, len(nodes))
		childMap := make(map[proto.AccessibilityAXNodeID][]proto.AccessibilityAXNodeID, len(nodes))
		for _, n := range nodes {
			nodeMap[n.NodeID] = n
			if len(n.ChildIDs) > 0 {
				childMap[n.NodeID] = n.ChildIDs
			}
		}
		return nodeMap, childMap
	}

	tests := []struct {
		name  string
		nodes []*proto.AccessibilityAXNode
		root  proto.AccessibilityAXNodeID
		want  string
	}{
		{
			name: "collects StaticText from direct children",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "li", Role: axVal("listitem"), ChildIDs: []proto.AccessibilityAXNodeID{"t1", "t2"}},
				{NodeID: "t1", ParentID: "li", Role: axVal("StaticText"), Name: axVal("John Doe")},
				{NodeID: "t2", ParentID: "li", Role: axVal("StaticText"), Name: axVal("Mar 10")},
			},
			root: "li",
			want: "John Doe | Mar 10",
		},
		{
			name: "skips interactive children entirely",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "li", Role: axVal("listitem"), ChildIDs: []proto.AccessibilityAXNodeID{"t1", "link1", "t2"}},
				{NodeID: "t1", ParentID: "li", Role: axVal("StaticText"), Name: axVal("John Doe")},
				{NodeID: "link1", ParentID: "li", Role: axVal("link"), Name: axVal("Meeting Tomorrow"), ChildIDs: []proto.AccessibilityAXNodeID{"linktext"}},
				{NodeID: "linktext", ParentID: "link1", Role: axVal("StaticText"), Name: axVal("Meeting Tomorrow")},
				{NodeID: "t2", ParentID: "li", Role: axVal("StaticText"), Name: axVal("Mar 10")},
			},
			root: "li",
			want: "John Doe | Mar 10",
		},
		{
			name: "recurses into generic wrappers",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "li", Role: axVal("listitem"), ChildIDs: []proto.AccessibilityAXNodeID{"div1"}},
				{NodeID: "div1", ParentID: "li", Role: axVal("generic"), ChildIDs: []proto.AccessibilityAXNodeID{"span1", "span2"}},
				{NodeID: "span1", ParentID: "div1", Role: axVal("generic"), ChildIDs: []proto.AccessibilityAXNodeID{"t1"}},
				{NodeID: "t1", ParentID: "span1", Role: axVal("StaticText"), Name: axVal("Sender")},
				{NodeID: "span2", ParentID: "div1", Role: axVal("generic"), ChildIDs: []proto.AccessibilityAXNodeID{"t2"}},
				{NodeID: "t2", ParentID: "span2", Role: axVal("StaticText"), Name: axVal("Date")},
			},
			root: "li",
			want: "Sender | Date",
		},
		{
			name: "skips single-char decorative text",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "li", Role: axVal("listitem"), ChildIDs: []proto.AccessibilityAXNodeID{"t1", "t2", "t3"}},
				{NodeID: "t1", ParentID: "li", Role: axVal("StaticText"), Name: axVal("Price")},
				{NodeID: "t2", ParentID: "li", Role: axVal("StaticText"), Name: axVal("·")},
				{NodeID: "t3", ParentID: "li", Role: axVal("StaticText"), Name: axVal("$29.99")},
			},
			root: "li",
			want: "Price | $29.99",
		},
		{
			name: "skips whitespace-only text",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "li", Role: axVal("listitem"), ChildIDs: []proto.AccessibilityAXNodeID{"t1", "t2", "t3"}},
				{NodeID: "t1", ParentID: "li", Role: axVal("StaticText"), Name: axVal("Hello")},
				{NodeID: "t2", ParentID: "li", Role: axVal("StaticText"), Name: axVal("   ")},
				{NodeID: "t3", ParentID: "li", Role: axVal("StaticText"), Name: axVal("World")},
			},
			root: "li",
			want: "Hello | World",
		},
		{
			name: "returns empty when all text is filtered out",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "li", Role: axVal("listitem"), ChildIDs: []proto.AccessibilityAXNodeID{"t1", "t2"}},
				{NodeID: "t1", ParentID: "li", Role: axVal("StaticText"), Name: axVal("·")},
				{NodeID: "t2", ParentID: "li", Role: axVal("StaticText"), Name: axVal(" ")},
			},
			root: "li",
			want: "",
		},
		{
			name: "returns empty when only interactive children",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "li", Role: axVal("listitem"), ChildIDs: []proto.AccessibilityAXNodeID{"btn1"}},
				{NodeID: "btn1", ParentID: "li", Role: axVal("button"), Name: axVal("Click me")},
			},
			root: "li",
			want: "",
		},
		{
			name: "truncates long summary",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "li", Role: axVal("listitem"), ChildIDs: []proto.AccessibilityAXNodeID{"t1", "t2"}},
				{NodeID: "t1", ParentID: "li", Role: axVal("StaticText"), Name: axVal(strings.Repeat("A", 100))},
				{NodeID: "t2", ParentID: "li", Role: axVal("StaticText"), Name: axVal(strings.Repeat("B", 100))},
			},
			root: "li",
			// "AAA...A | BBB...B" total > 160, truncated to 157 + "..."
			want: truncate(strings.Repeat("A", 100)+" | "+strings.Repeat("B", 100), maxTextLength),
		},
		{
			name: "skips buttons deep inside wrappers",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "art", Role: axVal("article"), ChildIDs: []proto.AccessibilityAXNodeID{"div1", "div2"}},
				{NodeID: "div1", ParentID: "art", Role: axVal("generic"), ChildIDs: []proto.AccessibilityAXNodeID{"t1", "btn1"}},
				{NodeID: "t1", ParentID: "div1", Role: axVal("StaticText"), Name: axVal("Product Name")},
				{NodeID: "btn1", ParentID: "div1", Role: axVal("button"), Name: axVal("Add to Cart")},
				{NodeID: "div2", ParentID: "art", Role: axVal("generic"), ChildIDs: []proto.AccessibilityAXNodeID{"t2"}},
				{NodeID: "t2", ParentID: "div2", Role: axVal("StaticText"), Name: axVal("$29.99")},
			},
			root: "art",
			want: "Product Name | $29.99",
		},
		{
			name: "handles missing child nodes gracefully",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "li", Role: axVal("listitem"), ChildIDs: []proto.AccessibilityAXNodeID{"t1", "missing", "t2"}},
				{NodeID: "t1", ParentID: "li", Role: axVal("StaticText"), Name: axVal("Hello")},
				{NodeID: "t2", ParentID: "li", Role: axVal("StaticText"), Name: axVal("World")},
			},
			root: "li",
			want: "Hello | World",
		},
		{
			name: "returns empty for no children",
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "li", Role: axVal("listitem")},
			},
			root: "li",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeMap, childMap := buildMaps(tt.nodes)
			got := collectContainerText(tt.root, nodeMap, childMap)
			if got != tt.want {
				t.Errorf("collectContainerText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContainerSummaryInSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		mode     SnapshotMode
		nodes    []*proto.AccessibilityAXNode
		contains []string
		excludes []string
	}{
		{
			name: "listitem gets synthetic summary in default mode",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"list1"}},
				{NodeID: "list1", ParentID: "root", Role: axVal("list"), BackendDOMNodeID: 300, ChildIDs: []proto.AccessibilityAXNodeID{"li1"}},
				{NodeID: "li1", ParentID: "list1", Role: axVal("listitem"), BackendDOMNodeID: 301, ChildIDs: []proto.AccessibilityAXNodeID{"t1", "cb1", "link1", "t2"}},
				{NodeID: "t1", ParentID: "li1", Role: axVal("StaticText"), Name: axVal("John Doe"), BackendDOMNodeID: 302},
				{NodeID: "cb1", ParentID: "li1", Role: axVal("checkbox"), Name: axVal("Select"), BackendDOMNodeID: 303},
				{NodeID: "link1", ParentID: "li1", Role: axVal("link"), Name: axVal("Meeting Tomorrow"), BackendDOMNodeID: 304},
				{NodeID: "t2", ParentID: "li1", Role: axVal("StaticText"), Name: axVal("Mar 10"), BackendDOMNodeID: 305},
			},
			contains: []string{
				`[301] listitem "John Doe | Mar 10"`,
				`    [303] checkbox "Select"`,
				`    [304] link "Meeting Tomorrow"`,
			},
		},
		{
			name: "named article keeps its real name, no synthetic summary",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"art1"}},
				{NodeID: "art1", ParentID: "root", Role: axVal("article"), Name: axVal("Blog Post"), BackendDOMNodeID: 310, ChildIDs: []proto.AccessibilityAXNodeID{"t1", "link1"}},
				{NodeID: "t1", ParentID: "art1", Role: axVal("StaticText"), Name: axVal("Some extra text"), BackendDOMNodeID: 311},
				{NodeID: "link1", ParentID: "art1", Role: axVal("link"), Name: axVal("Read more"), BackendDOMNodeID: 312},
			},
			contains: []string{
				`[310] article "Blog Post"`,
				`  [312] link "Read more"`,
			},
			excludes: []string{"Some extra text"},
		},
		{
			name: "interactive mode skips synthetic summary",
			mode: ModeInteractive,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"list1"}},
				{NodeID: "list1", ParentID: "root", Role: axVal("list"), BackendDOMNodeID: 320, ChildIDs: []proto.AccessibilityAXNodeID{"li1"}},
				{NodeID: "li1", ParentID: "list1", Role: axVal("listitem"), BackendDOMNodeID: 321, ChildIDs: []proto.AccessibilityAXNodeID{"t1", "link1"}},
				{NodeID: "t1", ParentID: "li1", Role: axVal("StaticText"), Name: axVal("John Doe"), BackendDOMNodeID: 322},
				{NodeID: "link1", ParentID: "li1", Role: axVal("link"), Name: axVal("Subject"), BackendDOMNodeID: 323},
			},
			contains: []string{
				"  [321] listitem\n",
				`    [323] link "Subject"`,
			},
			excludes: []string{"John Doe"},
		},
		{
			name: "all mode still gets synthetic summary",
			mode: ModeAll,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"list1"}},
				{NodeID: "list1", ParentID: "root", Role: axVal("list"), BackendDOMNodeID: 330, ChildIDs: []proto.AccessibilityAXNodeID{"li1"}},
				{NodeID: "li1", ParentID: "list1", Role: axVal("listitem"), BackendDOMNodeID: 331, ChildIDs: []proto.AccessibilityAXNodeID{"t1", "link1"}},
				{NodeID: "t1", ParentID: "li1", Role: axVal("StaticText"), Name: axVal("Sender Name"), BackendDOMNodeID: 332},
				{NodeID: "link1", ParentID: "li1", Role: axVal("link"), Name: axVal("Subject"), BackendDOMNodeID: 333},
			},
			contains: []string{
				`[331] listitem "Sender Name"`,
			},
		},
		{
			name: "non-content-container structural role does not get summary",
			mode: ModeDefault,
			nodes: []*proto.AccessibilityAXNode{
				{NodeID: "root", Role: axVal("RootWebArea"), ChildIDs: []proto.AccessibilityAXNodeID{"nav1"}},
				{NodeID: "nav1", ParentID: "root", Role: axVal("navigation"), BackendDOMNodeID: 340, ChildIDs: []proto.AccessibilityAXNodeID{"t1", "link1"}},
				{NodeID: "t1", ParentID: "nav1", Role: axVal("StaticText"), Name: axVal("Menu Label"), BackendDOMNodeID: 341},
				{NodeID: "link1", ParentID: "nav1", Role: axVal("link"), Name: axVal("Home"), BackendDOMNodeID: 342},
			},
			contains: []string{
				"[340] navigation\n",
				`  [342] link "Home"`,
			},
			excludes: []string{"Menu Label"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSnapshot("https://example.com", "Test", tt.nodes, tt.mode)

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
