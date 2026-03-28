//go:build integration

// Package integration contains browser-level integration tests that run against
// a real Chrome instance and a local HTTP server.
//
// Adding a new test:
//  1. Define an HTML constant (e.g. const myPageHTML = `...`) in a new or existing test file.
//  2. Register it in the pages map below (e.g. "/my-page": myPageHTML).
//  3. Write tests that call b.Open(ctx, pageURL("/my-page"), false) and use
//     b.Snapshot, b.Click, etc. directly.
//
// Run with: go test -tags integration ./test/integration/ -v -count=1
package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/usewebtool/webtool/browser"
)

var (
	b      *browser.Browser
	server *httptest.Server
)

const integrationTestTimeout = 30 * time.Second

// pages maps route paths to HTML content. Add new fixtures here.
var pages = map[string]string{
	"/simple":     simpleHTML,
	"/controlled": controlledHTML,
	"/spa":        spaHTML,
	"/extract":    extractHTML,
	"/dynamic":    dynamicHTML,
}

func TestMain(m *testing.M) {
	// Start local HTTP server serving embedded HTML fixtures.
	mux := http.NewServeMux()
	for path, content := range pages {
		body := content // capture for closure
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, body)
		})
	}
	server = httptest.NewServer(mux)

	// Connect to Chrome once for all tests.
	b = browser.New()
	if err := b.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to Chrome: %v\n", err)
		server.Close()
		os.Exit(1)
	}

	code := m.Run()

	b.Close()
	server.Close()
	os.Exit(code)
}

// pageURL returns the full URL for a fixture path (e.g. "/simple").
func pageURL(path string) string {
	return server.URL + path
}

// findElement returns the backendNodeId string for the first element matching
// the given substring in a snapshot string. Snapshot lines look like: [12345] role "name"
func findElement(t *testing.T, snapshot, match string) string {
	t.Helper()
	for _, line := range strings.Split(snapshot, "\n") {
		if strings.Contains(line, match) {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "[") {
				continue
			}
			end := strings.Index(line, "]")
			if end < 0 {
				continue
			}
			return line[1:end]
		}
	}
	t.Fatalf("no element matching %q found in snapshot:\n%s", match, snapshot)
	return ""
}

func findTabByURL(t *testing.T, tabs []browser.TabInfo, url string) browser.TabInfo {
	t.Helper()
	want := strings.SplitN(url, "#", 2)[0]
	for _, tab := range tabs {
		got := strings.SplitN(tab.URL, "#", 2)[0]
		if got == want {
			return tab
		}
	}
	t.Fatalf("no tab with URL %q found in tabs: %+v", url, tabs)
	return browser.TabInfo{}
}

const simpleHTML = `<!DOCTYPE html>
<html>
<head><title>Simple Test</title></head>
<body>
	<h1>Hello</h1>
	<button id="btn" onclick="document.getElementById('output').textContent = 'clicked'">Click me</button>
	<div id="output"></div>
</body>
</html>`

const controlledHTML = `<!DOCTYPE html>
<html>
<head>
	<title>Controlled Form</title>
	<meta charset="utf-8">
</head>
<body>
	<div id="app"></div>
	<script>
	(() => {
		const state = {
			draft: "",
			priority: "Normal",
			notes: "",
			ownerEmail: "",
			deliveryMode: "Standard",
			notifyTeam: false,
			fileName: "",
			items: [],
			status: "Idle"
		};

		function summaryText() {
			return "Task: " + (state.draft || "Nothing yet") +
				" | Priority: " + state.priority +
				" | Notes: " + (state.notes || "None") +
				" | Owner: " + (state.ownerEmail || "Unassigned") +
				" | Delivery: " + state.deliveryMode +
				" | Notify: " + (state.notifyTeam ? "Yes" : "No");
		}

		function render() {
			const itemsHTML = state.items.map(item => "<li>" + escapeHTML(item) + "</li>").join("");
			document.getElementById("app").innerHTML =
				"<main>" +
					"<h1>Task board</h1>" +
					"<label for=\"priority-select\">Priority</label>" +
					"<select id=\"priority-select\">" +
						"<option" + (state.priority === "Low" ? " selected" : "") + ">Low</option>" +
						"<option" + (state.priority === "Normal" ? " selected" : "") + ">Normal</option>" +
						"<option" + (state.priority === "Urgent" ? " selected" : "") + ">Urgent</option>" +
					"</select>" +
					"<label for=\"task-input\">Task name</label>" +
					"<input id=\"task-input\" type=\"text\" value=\"" + escapeHTML(state.draft) + "\" autocomplete=\"off\">" +
					"<label for=\"notes-input\">Notes</label>" +
					"<textarea id=\"notes-input\">" + escapeHTML(state.notes) + "</textarea>" +
					"<label for=\"owner-email\">Owner email</label>" +
					"<input id=\"owner-email\" type=\"email\" value=\"" + escapeHTML(state.ownerEmail) + "\" autocomplete=\"off\">" +
					"<fieldset>" +
						"<legend>Delivery mode</legend>" +
						"<label><input type=\"radio\" name=\"delivery-mode\" value=\"Standard\"" + (state.deliveryMode === "Standard" ? " checked" : "") + ">Standard delivery</label>" +
						"<label><input type=\"radio\" name=\"delivery-mode\" value=\"Expedite\"" + (state.deliveryMode === "Expedite" ? " checked" : "") + ">Expedite delivery</label>" +
					"</fieldset>" +
					"<label><input id=\"notify-team\" type=\"checkbox\"" + (state.notifyTeam ? " checked" : "") + ">Notify team</label>" +
					"<label for=\"attachment\">Attachment</label>" +
					"<input id=\"attachment\" type=\"file\">" +
					"<div id=\"file-status\">" + escapeHTML(state.fileName || "No file selected") + "</div>" +
					"<button id=\"save-btn\" " + (state.draft.trim() ? "" : "disabled") + ">Add task</button>" +
					"<div role=\"status\" aria-live=\"polite\">" + escapeHTML(state.status) + "</div>" +
					"<p>Preview: " + escapeHTML(summaryText()) + "</p>" +
					"<ul>" + itemsHTML + "</ul>" +
				"</main>";
		}

		function submit() {
			const value = state.draft.trim();
			if (!value) return;
			const item = "[" + state.priority + "] " + value +
				" / " + state.deliveryMode +
				" / notify=" + (state.notifyTeam ? "yes" : "no") +
				" / owner=" + (state.ownerEmail || "none") +
				" / notes=" + (state.notes || "none");
			state.items.push(item);
			state.status = "Added " + item;
			state.draft = "";
			state.notes = "";
			state.ownerEmail = "";
			state.deliveryMode = "Standard";
			state.notifyTeam = false;
			render();
			document.getElementById("task-input").focus();
		}

		function escapeHTML(value) {
			return value
				.replace(/&/g, "&amp;")
				.replace(/</g, "&lt;")
				.replace(/>/g, "&gt;")
				.replace(/"/g, "&quot;");
		}

		document.addEventListener("input", (event) => {
			if (event.target.id === "task-input") {
				state.draft = event.target.value;
				state.status = state.draft ? "Draft ready" : "Idle";
			} else if (event.target.id === "notes-input") {
				state.notes = event.target.value;
				state.status = state.notes ? "Notes updated" : "Notes cleared";
			} else if (event.target.id === "owner-email") {
				state.ownerEmail = event.target.value;
				state.status = state.ownerEmail ? "Owner updated" : "Owner cleared";
			} else {
				return;
			}
			render();
			if (event.target.id === "task-input") {
				document.getElementById("task-input").focus();
			}
			if (event.target.id === "notes-input") {
				document.getElementById("notes-input").focus();
			}
			if (event.target.id === "owner-email") {
				document.getElementById("owner-email").focus();
			}
		});

		document.addEventListener("change", (event) => {
			if (event.target.id === "priority-select") {
				state.priority = event.target.value;
				state.status = "Priority set to " + state.priority;
			} else if (event.target.name === "delivery-mode") {
				state.deliveryMode = event.target.value;
				state.status = "Delivery set to " + state.deliveryMode;
			} else if (event.target.id === "notify-team") {
				state.notifyTeam = event.target.checked;
				state.status = state.notifyTeam ? "Notifications enabled" : "Notifications disabled";
			} else if (event.target.id === "attachment") {
				state.fileName = event.target.files.length > 0 ? event.target.files[0].name : "";
				state.status = state.fileName ? "File attached: " + state.fileName : "File removed";
			} else {
				return;
			}
			render();
			if (event.target.id === "priority-select") {
				document.getElementById("priority-select").focus();
			}
		});

		document.addEventListener("click", (event) => {
			if (event.target.id === "save-btn") {
				submit();
			}
		});

		document.addEventListener("keydown", (event) => {
			if (event.target.id === "task-input" && event.key === "Enter") {
				event.preventDefault();
				submit();
			}
		});

		render();
	})();
	</script>
</body>
</html>`

const dynamicHTML = `<!DOCTYPE html>
<html>
<head>
	<title>Dynamic Test</title>
	<meta charset="utf-8">
	<style>
		.card-actions { display: none; }
		.card:hover .card-actions { display: block; }
		#app { min-height: 200vh; }
	</style>
</head>
<body>
	<div id="app">
		<h1>Dashboard</h1>

		<!-- Delayed element: appears after 500ms (for Wait test) -->
		<div id="notifications"></div>

		<!-- Hover reveal: actions visible only on hover (for Hover test) -->
		<div class="card" id="project-card">
			<h2>Project Alpha</h2>
			<p>Status: active</p>
			<div class="card-actions">
				<button id="archive-btn">Archive</button>
				<button id="delete-btn">Delete</button>
			</div>
		</div>

		<!-- Re-render section: clicking refresh destroys and recreates nodes (for Stale node test) -->
		<div id="feed">
			<h2>Activity Feed</h2>
			<ul id="feed-list">
				<li id="feed-item-1">Deploy v1.0 completed</li>
				<li id="feed-item-2">Test suite passed</li>
				<li id="feed-item-3">PR #42 merged</li>
			</ul>
			<button id="refresh-feed">Refresh Feed</button>
		</div>

		<!-- Shadow DOM: web component with encapsulated internals (for Shadow DOM test) -->
		<shadow-card></shadow-card>

		<!-- Counter: changes on interaction, resets on reload (for Reload test) -->
		<div id="counter-section">
			<span id="counter-value">0</span>
			<button id="increment-btn">Increment</button>
		</div>
	</div>

	<script>
	(() => {
		// Shadow DOM web component.
		class ShadowCard extends HTMLElement {
			connectedCallback() {
				const shadow = this.attachShadow({ mode: "open" });
				shadow.innerHTML =
					'<div>' +
						'<h2>Shadow Heading</h2>' +
						'<p id="shadow-status">Idle</p>' +
						'<button id="shadow-btn">Shadow Action</button>' +
					'</div>';
				shadow.getElementById("shadow-btn").addEventListener("click", () => {
					shadow.getElementById("shadow-status").textContent = "Shadow clicked";
				});
			}
		}
		customElements.define("shadow-card", ShadowCard);

		// Delayed notification appears after 2s — long enough that page load
		// and waitPageSettle have already completed, so Wait must actually poll.
		setTimeout(() => {
			document.getElementById("notifications").innerHTML =
				'<div id="delayed-notification" role="alert">New deployment ready</div>';
		}, 2000);

		// Refresh feed: destroys all existing list items and creates new ones.
		document.getElementById("refresh-feed").addEventListener("click", () => {
			const list = document.getElementById("feed-list");
			list.innerHTML =
				'<li id="feed-item-4">Hotfix v1.0.1 deployed</li>' +
				'<li id="feed-item-5">Monitoring alert cleared</li>';
		});

		// Archive button (revealed on hover).
		document.addEventListener("click", (e) => {
			if (e.target.id === "archive-btn") {
				const card = document.getElementById("project-card");
				card.querySelector("p").textContent = "Status: archived";
			}
		});

		// Counter.
		let count = 0;
		document.getElementById("increment-btn").addEventListener("click", () => {
			count++;
			document.getElementById("counter-value").textContent = String(count);
		});
	})();
	</script>
</body>
</html>`

const extractHTML = `<!DOCTYPE html>
<html>
<head><title>Extract Test</title><meta charset="utf-8"></head>
<body>
	<header><nav><a href="/home">Home</a></nav></header>
	<main>
		<h1>Main Heading</h1>
		<p>This is the main content paragraph.</p>
		<a href="https://example.com">Example Link</a>
		<ul>
			<li>Item one</li>
			<li>Item two</li>
		</ul>
	</main>
	<footer><p>Footer content here</p></footer>
</body>
</html>`

const spaHTML = `<!DOCTYPE html>
<html>
<head>
	<title>Single Page App</title>
	<meta charset="utf-8">
</head>
<body>
	<div id="app"></div>
	<script>
	(() => {
		const state = {
			route: "home"
		};

		function setRoute(route, push) {
			state.route = route;
			if (push) {
				history.pushState({route}, "", "#" + route);
			}
			render();
		}

		function routeBody() {
			if (state.route === "settings") {
				return "<section>" +
					"<h1>Settings</h1>" +
					"<p>Notifications are enabled.</p>" +
					"<button id=\"save-settings\">Save preferences</button>" +
				"</section>";
			}

			return "<section>" +
				"<h1>Home</h1>" +
				"<p>Dashboard overview.</p>" +
			"</section>";
		}

		function render() {
			document.getElementById("app").innerHTML =
				"<main>" +
					"<nav aria-label=\"Primary\">" +
						"<button id=\"nav-home\">Home</button>" +
						"<button id=\"nav-settings\">Settings</button>" +
					"</nav>" +
					"<div role=\"status\" aria-live=\"polite\">Route: " + state.route + "</div>" +
					routeBody() +
				"</main>";
		}

		document.addEventListener("click", (event) => {
			if (event.target.id === "nav-home") {
				setRoute("home", true);
			}
			if (event.target.id === "nav-settings") {
				setRoute("settings", true);
			}
		});

		window.addEventListener("popstate", (event) => {
			const route = event.state && event.state.route ? event.state.route : "home";
			setRoute(route, false);
		});

		if (location.hash === "#settings") {
			state.route = "settings";
		}
		history.replaceState({route: state.route}, "", "#" + state.route);
		render();
	})();
	</script>
</body>
</html>`
