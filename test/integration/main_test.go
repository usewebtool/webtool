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
